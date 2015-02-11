package main

import (
	"sync"
	"time"

	"github.com/miekg/dns"
)

type Question struct {
	qname  string
	qtype  string
	qclass string
}

const (
	notIPQuery = 0
	_IP4Query  = 4
	_IP6Query  = 6
)

func (q *Question) String() string {
	return q.qname + " " + q.qclass + " " + q.qtype
}

type GODNSHandler struct {
	resolver *Resolver
	cache    Cache
	hosts    Hosts
	mu       *sync.Mutex
}

func NewHandler() *GODNSHandler {

	var (
		clientConfig *dns.ClientConfig
		cacheConfig  CacheSettings
		resolver     *Resolver
		cache        Cache
	)

	resolvConfig := settings.ResolvConfig
	clientConfig, err := dns.ClientConfigFromFile(resolvConfig.ResolvFile)
	if err != nil {
		logger.Printf(":%s is not a valid resolv.conf file\n", resolvConfig.ResolvFile)
		logger.Println(err)
		panic(err)
	}
	clientConfig.Timeout = resolvConfig.Timeout
	resolver = &Resolver{clientConfig}

	cacheConfig = settings.Cache
	switch cacheConfig.Backend {
	case "memory":
		cache = &MemoryCache{
			Backend:  make(map[string]Mesg),
			Expire:   time.Duration(cacheConfig.Expire) * time.Second,
			Maxcount: cacheConfig.Maxcount,
			mu:       new(sync.RWMutex),
		}
	case "redis":
		// cache = &MemoryCache{
		// 	Backend:    make(map[string]*dns.Msg),
		//  Expire:   time.Duration(cacheConfig.Expire) * time.Second,
		// 	Serializer: new(JsonSerializer),
		// 	Maxcount:   cacheConfig.Maxcount,
		// }
		panic("Redis cache backend not implement yet")
	default:
		logger.Printf("Invalid cache backend %s", cacheConfig.Backend)
		panic("Invalid cache backend")
	}

	hosts := NewHosts(settings.Hosts, settings.Redis)

	return &GODNSHandler{resolver, cache, hosts, new(sync.Mutex)}
}

func (h *GODNSHandler) do(Net string, w dns.ResponseWriter, req *dns.Msg) {
	q := req.Question[0]
	Q := Question{UnFqdn(q.Name), dns.TypeToString[q.Qtype], dns.ClassToString[q.Qclass]}

	Debug("Question:　%s", Q.String())

	IPQuery := h.isIPQuery(q)

	// Query hosts
	if settings.Hosts.Enable && IPQuery > 0 {
		if ip, ok := h.hosts.Get(Q.qname, IPQuery); ok {
			m := new(dns.Msg)
			m.SetReply(req)

			switch IPQuery {
			case _IP4Query:
				rr_header := dns.RR_Header{
					Name:   q.Name,
					Rrtype: dns.TypeA,
					Class:  dns.ClassINET,
					Ttl:    settings.Hosts.TTL,
				}
				a := &dns.A{rr_header, ip}
				m.Answer = append(m.Answer, a)
			case _IP6Query:
				rr_header := dns.RR_Header{
					Name:   q.Name,
					Rrtype: dns.TypeAAAA,
					Class:  dns.ClassINET,
					Ttl:    settings.Hosts.TTL,
				}
				aaaa := &dns.AAAA{rr_header, ip}
				m.Answer = append(m.Answer, aaaa)
			}

			w.WriteMsg(m)
			Debug("%s found in hosts file", Q.qname)
			return
		} else {
			Debug("%s didn't found in hosts file", Q.qname)
		}

	}

	// Only query cache when qtype == 'A' , qclass == 'IN'
	var key string
	if IPQuery > 0 {
		key = KeyGen(Q)
		mesg, err := h.cache.Get(key)
		if err != nil {
			Debug("%s didn't hit cache: %s", Q.String(), err)
		} else {
			Debug("%s hit cache", Q.String())
			h.mu.Lock()
			mesg.Id = req.Id
			w.WriteMsg(mesg)
			h.mu.Unlock()
			return
		}

	}

	// Query on both tcp and udp, simultaneously.
	nextNet := "tcp"
	if Net == "tcp" {
		nextNet = "udp"
	}
	res := make(chan *dns.Msg, 1)
	errch := make(chan error, 1)

	L := func(net string) {
		msg, err := h.resolver.Lookup(net, req)
		if err != nil {
			errch <- err
			return
		}
		res <- msg
	}

	// Start asking on Net
	go L(Net)

	var (
		msg *dns.Msg
		err error
	)
	select {
	case msg = <-res:
	case err = <-errch:
	case <-time.After(1 * time.Second):
	}

	if err != nil || msg == nil {
		// after 1 second, or error, start on nextNet
		L(nextNet)
		select {
		case msg = <-res:
		case err = <-errch:
		}
	}

	if err != nil {
		Debug("%s", err)
		dns.HandleFailed(w, req)
		return
	}

	w.WriteMsg(msg)

	if key != "" && len(msg.Answer) > 0 {
		if err := h.cache.Set(key, msg); err != nil {
			Debug("Set %s cache failed: %s", Q.String(), err.Error())
		}

		Debug("Insert %s into cache", Q.String())
	}

}

func (h *GODNSHandler) DoTCP(w dns.ResponseWriter, req *dns.Msg) {
	h.do("tcp", w, req)
}

func (h *GODNSHandler) DoUDP(w dns.ResponseWriter, req *dns.Msg) {
	h.do("udp", w, req)
}

func (h *GODNSHandler) isIPQuery(q dns.Question) int {
	if q.Qclass != dns.ClassINET {
		return notIPQuery
	}

	switch q.Qtype {
	case dns.TypeA:
		return _IP4Query
	case dns.TypeAAAA:
		return _IP6Query
	default:
		return notIPQuery
	}
}

func UnFqdn(s string) string {
	if dns.IsFqdn(s) {
		return s[:len(s)-1]
	}
	return s
}
