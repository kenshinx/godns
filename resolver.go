package main

import (
	"bufio"
	"fmt"
	"os"
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
	config        *dns.ClientConfig
	domain_server *suffixTreeNode
}

func NewResolver(c ResolvSettings) *Resolver {
	var clientConfig *dns.ClientConfig
	clientConfig, err := dns.ClientConfigFromFile(c.ResolvFile)
	if err != nil {
		logger.Printf(":%s is not a valid resolv.conf file\n", c.ResolvFile)
		logger.Println(err)
		panic(err)
	}
	clientConfig.Timeout = c.Timeout

	domain_server := newSuffixTreeRoot()
	r := &Resolver{clientConfig, domain_server}

	if len(c.DomainServerFile) > 0 {
		r.ReadDomainServerFile(c.DomainServerFile)
	}
	return r
}

func (r *Resolver) ReadDomainServerFile(file string) {
	buf, err := os.Open(file)
	if err != nil {
		panic("Can't open " + file)
	}
	scanner := bufio.NewScanner(buf)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		if !strings.HasPrefix(line, "server") {
			continue
		}

		sli := strings.Split(line, "=")
		if len(sli) != 2 {
			continue
		}

		line = strings.TrimSpace(sli[1])

		tokens := strings.Split(line, "/")
		if len(tokens) != 3 {
			continue
		}
		domain := tokens[1]
		ip := tokens[2]
		if !isDomain(domain) || !isIP(ip) {
			continue
		}
		r.domain_server.sinsert(strings.Split(domain, "."), ip)
	}

}

func (r *Resolver) Lookup(net string, req *dns.Msg) (message *dns.Msg, err error) {
	c := &dns.Client{
		Net:          net,
		ReadTimeout:  r.Timeout(),
		WriteTimeout: r.Timeout(),
	}

	qname := req.Question[0].Name
	nameservers := r.Nameservers(qname)
	for _, nameserver := range nameservers {
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
		Debug("%s resolv on %s rtt: %v", UnFqdn(qname), nameserver, rtt)
		return r, nil
	}
	return nil, ResolvError{qname, nameservers}
}

func (r *Resolver) Nameservers(qname string) []string {

	queryKeys := strings.Split(qname, ".")
	queryKeys = queryKeys[:len(queryKeys)-1] // ignore last '.'

	ns := []string{}
	if v, found := r.domain_server.search(queryKeys); found {
		Debug("found upstream: %v", v)
		server := v
		nameserver := server + ":53"
		ns = append(ns, nameserver)
	}

	for _, server := range r.config.Servers {
		nameserver := server + ":" + r.config.Port
		ns = append(ns, nameserver)
	}
	return ns
}

func (r *Resolver) Timeout() time.Duration {
	return time.Duration(r.config.Timeout) * time.Second
}
