package main

import (
	"github.com/miekg/dns"
	// "log"
	"fmt"
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
			backend:    make(map[string]string),
			serializer: new(JsonSerializer),
			expire:     cacheConfig.Expire,
			maxcount:   cacheConfig.Maxcount,
		}
	case "redis":
		cache = &MemoryCache{
			backend:    make(map[string]string),
			serializer: new(JsonSerializer),
			expire:     cacheConfig.Expire,
			maxcount:   cacheConfig.Maxcount,
		}
	default:
		logger.Printf("Invalid cache backend %s", cacheConfig.Backend)
		panic("Invalid cache backend")
	}
	return &GODNSHandler{resolver, cache}
}

func (h *GODNSHandler) do(net string, w dns.ResponseWriter, req *dns.Msg) {

	q := req.Question[0]
	Q := Question{q.Name, dns.TypeToString[q.Qtype], dns.ClassToString[q.Qclass]}

	Debug("Question:　%s", Q.String())

	key := KeyGen(Q)
	fmt.Println(key)
	// Only query cache when qtype == 'A' , qclass == 'IN'
	if q.Qtype == dns.TypeA && q.Qclass == dns.ClassINET {
		mesg, err := h.cache.Get(key)
		if err != nil {
			Debug("%s didn't hit cache: %s", Q.String(), err)
		} else {
			Debug("%s hit cache", Q.String())
			fmt.Println(string(mesg))
			w.Write(mesg)
			return
		}

	}

	mesg, err := h.resolver.Lookup(net, req)

	if err != nil {
		Debug("%s", err)
		dns.HandleFailed(w, req)
		return
	}

	w.WriteMsg(mesg)

	if q.Qtype == dns.TypeA && q.Qclass == dns.ClassINET {
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
