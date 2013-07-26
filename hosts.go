package main

import (
	"github.com/hoisie/redis"
	// "github.com/miekg/dns"
	"fmt"
	"io/ioutil"
)

type HostsQueryFaild struct {
	domain string
}

func (e HostsQueryFaild) Error() string {
	return e.domain + " hosts match failed"
}

func readLocalHostsFile(file string) map[string]string {
	var hosts = make(map[string]string)
	f, _ := os.Open(file)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {

		line := scanner.Text()
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		sli := strings.Split(line, " ")
		if len(sli) == 1 {
			sli = strings.Split(line, "\t")
		}

		if len(sli) < 2 {
			continue
		}

		domain := sli[len(sli)-1]
		ip := sli[0]
		if !isDomain(domain) || !isIP(ip) {
			continue
		}

		hosts[domain] = ip
	}
	return hosts
}

func isDomain(domain string) bool {
	match, _ := regexp.MatchString("^[a-z]", domain)
	return match
}

func isIP(ip string) bool {
	match, _ := regexp.MatchString("^[1-9]", ip)
	return match
}
