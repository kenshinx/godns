package main

import (
	"github.com/miekg/dns"
	// "log"
	// "fmt"
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

	q := req.Question[0]
	Q := Question{q.Name, dns.TypeToString[q.Qtype], dns.ClassToString[q.Qclass]}

	Debug("Question:　%s", Q.String())

	mesg, err := h.resolver.Lookup(net, req)

	if err != nil {
		Debug("%s", err)
		dns.HandleFailed(w, req)
		return
	}

	w.WriteMsg(mesg)

}

func (h *GODNSHandler) DoTCP(w dns.ResponseWriter, req *dns.Msg) {
	h.do("tcp", w, req)
}

func (h *GODNSHandler) DoUDP(w dns.ResponseWriter, req *dns.Msg) {
	h.do("udp", w, req)
}

type Question struct {
	qname  string
	qtype  string
	qclass string
}

func (q *Question) String() string {
	return q.qname + " " + q.qclass + " " + q.qtype
}
