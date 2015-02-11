package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/miekg/dns"
)

type ResolvError struct {
	qname       string
	nameservers []string
}

func (e ResolvError) Error() string {
	errmsg := fmt.Sprintf("%s resolv failed on %s", e.qname, strings.Join(e.nameservers, "; "))
	return errmsg
}

type Resolver struct {
	config *dns.ClientConfig
}

func (r *Resolver) Lookup(net string, req *dns.Msg) (message *dns.Msg, err error) {
	c := &dns.Client{
		Net:          net,
		ReadTimeout:  r.Timeout(),
		WriteTimeout: r.Timeout(),
	}

	qname := req.Question[0].Name

	for _, nameserver := range r.Nameservers() {
		r, rtt, err := c.Exchange(req, nameserver)
		if err != nil {
			Debug("%s socket error on %s", qname, nameserver)
			Debug("error:%s", err.Error())
			continue
		}
		if r != nil && r.Rcode != dns.RcodeSuccess {
			Debug("%s failed to get an valid answer on %s", qname, nameserver)
			continue
		}
		Debug("%s resolv on %s ttl: %d", UnFqdn(qname), nameserver, rtt)
		return r, nil
	}
	return nil, ResolvError{qname, r.Nameservers()}

}

// Namservers return the array of nameservers, with port number appended.
// '#' in the name is treated as port separator, as with dnsmasq.
func (r *Resolver) Nameservers() (ns []string) {
	for _, server := range r.config.Servers {
		if i := strings.IndexByte(server, '#'); i > 0 {
			server = server[:i] + ":" + server[i+1:]
		} else {
			server = server + ":" + r.config.Port
		}
		ns = append(ns, server)
	}
	return
}

func (r *Resolver) Timeout() time.Duration {
	return time.Duration(r.config.Timeout) * time.Second
}
