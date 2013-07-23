package main

import (
	"fmt"
	"github.com/miekg/dns"
	"time"
)

type Resolver struct {
	config *dns.ClientConfig
}

func (r *Resolver) Lookup(net string, req *dns.Msg) {
	c := &dns.Client{
		Net:          net,
		ReadTimeout:  r.Timeout(),
		WriteTimeout: r.Timeout(),
	}

	for _, nameserver := range r.Nameservers() {
		r, rtt, _ := c.Exchange(req, nameserver)
		fmt.Println(r)
		fmt.Println(rtt)

	}
	// r,rtt,_:c.Exchange(req, a)

}

func (r *Resolver) Nameservers() (ns []string) {
	for _, server := range r.config.Servers {
		nameserver := server + ":" + r.config.Port
		ns = append(ns, nameserver)
	}
	return
}

func (r *Resolver) Timeout() time.Duration {
	return time.Duration(r.config.Timeout) * time.Second
}
