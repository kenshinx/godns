package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
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
	servers       []string
	domain_server *suffixTreeNode
	config        *ResolvSettings
}

func NewResolver(c ResolvSettings) *Resolver {
	r := &Resolver{
		servers:       []string{},
		domain_server: newSuffixTreeRoot(),
		config:        &c,
	}

	if len(c.ServerListFile) > 0 {
		r.ReadServerListFile(c.ServerListFile)
		// Debug("%v", r.servers)
	}

	if len(c.ResolvFile) > 0 {
		clientConfig, err := dns.ClientConfigFromFile(c.ResolvFile)
		if err != nil {
			logger.Printf(":%s is not a valid resolv.conf file\n", c.ResolvFile)
			logger.Println(err)
			panic(err)
		}
		for _, server := range clientConfig.Servers {
			nameserver := server + ":" + clientConfig.Port
			r.servers = append(r.servers, nameserver)
		}
	}

	return r
}

func (r *Resolver) ReadServerListFile(file string) {
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
		switch len(tokens) {
		case 3:
			domain := tokens[1]
			ip := tokens[2]
			if !isDomain(domain) || !isIP(ip) {
				continue
			}
			r.domain_server.sinsert(strings.Split(domain, "."), ip)
		case 1:
			srv_port := strings.Split(line, "#")
			if len(srv_port) > 2 {
				continue
			}

			ip := ""
			if ip = srv_port[0]; !isIP(ip) {
				continue
			}

			port := "53"
			if len(srv_port) == 2 {
				if _, err := strconv.Atoi(srv_port[1]); err != nil {
					continue
				}
				port = srv_port[1]
			}
			r.servers = append(r.servers, ip+":"+port)

		}
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

	for _, nameserver := range r.servers {
		ns = append(ns, nameserver)
	}
	return ns
}

func (r *Resolver) Timeout() time.Duration {
	return time.Duration(r.config.Timeout) * time.Second
}
