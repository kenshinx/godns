package main

import (
	"github.com/miekg/dns"
	"net"
	"sync"
	"time"
)

type Question struct {
	qname  string
	qtype  string
	qclass string
}

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

	// Query hosts
	if settings.Hosts.Enable && h.isIPQuery(q) {
		if ip, ok := h.hosts.Get(Q.qname); ok {
			m := new(dns.Msg)
			m.SetReply(req)
			rr_header := dns.RR_Header{Name: q.Name, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: settings.Hosts.TTL}
			a := &dns.A{rr_header, net.ParseIP(ip)}
			m.Answer = append(m.Answer, a)
			w.WriteMsg(m)
			Debug("%s found in hosts", Q.qname)
			return
		}

	}

	// Only query cache when qtype == 'A' , qclass == 'IN'
	key := KeyGen(Q)
	if h.isIPQuery(q) {
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

	mesg, err := h.resolver.Lookup(Net, req)

	if err != nil {
		Debug("%s", err)
		dns.HandleFailed(w, req)
		return
	}

	w.WriteMsg(mesg)

	if h.isIPQuery(q) {
		err = h.cache.Set(key, mesg)

		if err != nil {
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

func (h *GODNSHandler) isIPQuery(q dns.Question) bool {
	return q.Qtype == dns.TypeA && q.Qclass == dns.ClassINET
}

func UnFqdn(s string) string {
	if dns.IsFqdn(s) {
		return s[:len(s)-1]
	}
	return s
}
