package main

import (
	"github.com/miekg/dns"
	// "log"
)

type GODNSHandler struct {
	resolver *Resolver
	cache    *Cache
}

func NewHandler() *GODNSHandler {

	var (
		clientConfig *dns.ClientConfig
		cacheConfig  CacheSettings
	)

	resolvConfig := settings.ResolvConfig
	clientConfig, err := dns.ClientConfigFromFile(resolvConfig.ResolvFile)
	if err != nil {
		logger.Printf(":%s is not a valid resolv.conf file\n", resolvConfig.ResolvFile)
		logger.Println(err)
		panic(err)
	}
	clientConfig.Timeout = resolvConfig.Timeout
	resolver := &Resolver{clientConfig}

	cacheConfig = settings.Cache
	cache := &Cache{cacheConfig}

	return &GODNSHandler{resolver, cache}
}

func (h *GODNSHandler) do(net string, w dns.ResponseWriter, req *dns.Msg) {

	qname := req.Question[0].Name
	qtype := req.Question[0].Qtype
	qclass := req.Question[0].Qclass

	Debug("Question:　%s %s %s", qname, dns.ClassToString[qclass], dns.TypeToString[qtype])

	h.resolver.Lookup(net, req)

}

func (h *GODNSHandler) DoTCP(w dns.ResponseWriter, req *dns.Msg) {
	h.do("tcp", w, req)
}

func (h *GODNSHandler) DoUDP(w dns.ResponseWriter, req *dns.Msg) {
	h.do("udp", w, req)
}
