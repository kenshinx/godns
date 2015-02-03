package main

import (
	"bufio"
	"net"
	"os"
	"regexp"
	"strings"

	"github.com/hoisie/redis"
)

type Hosts struct {
	FileHosts  map[string]string
	RedisHosts *RedisHosts
}

func NewHosts(hs HostsSettings, rs RedisSettings) Hosts {
	fileHosts := &FileHosts{hs.HostsFile}
	redis := &redis.Client{Addr: rs.Addr(), Db: rs.DB, Password: rs.Password}
	redisHosts := &RedisHosts{redis, hs.RedisKey}

	hosts := Hosts{fileHosts.GetAll(), redisHosts}
	return hosts

}

/*
1. Resolve hosts file only one times
2. Request redis on every query called, not found performance lose serious yet.
3. Match local /etc/hosts file first, remote redis records second
*/

func (h *Hosts) Get(domain string, family int) (ip net.IP, ok bool) {
	var sip string

	if sip, ok = h.FileHosts[domain]; !ok {
		if sip, ok = h.RedisHosts.Get(domain); !ok {
			return nil, false
		}
	}

	switch family {
	case _IP4Query:
		ip = net.ParseIP(sip).To4()
		return ip, (ip != nil)
	case _IP6Query:
		ip = net.ParseIP(sip).To16()
		return ip, (ip != nil)
	}
	return nil, false
}

func (h *Hosts) GetAll() map[string]string {

	m := make(map[string]string)
	for domain, ip := range h.RedisHosts.GetAll() {
		m[domain] = ip
	}
	for domain, ip := range h.FileHosts {
		m[domain] = ip
	}
	return m
}

type RedisHosts struct {
	redis *redis.Client
	key   string
}

func (r *RedisHosts) GetAll() map[string]string {
	var hosts = make(map[string]string)
	r.redis.Hgetall(r.key, hosts)
	return hosts
}

func (r *RedisHosts) Get(domain string) (ip string, ok bool) {
	b, err := r.redis.Hget(r.key, domain)
	return string(b), err == nil
}

func (r *RedisHosts) Set(domain, ip string) (bool, error) {
	return r.redis.Hset(r.key, domain, []byte(ip))
}

type FileHosts struct {
	file string
}

func (f *FileHosts) GetAll() map[string]string {
	var hosts = make(map[string]string)

	buf, err := os.Open(f.file)
	if err != nil {
		panic("Can't open " + f.file)
	}

	scanner := bufio.NewScanner(buf)
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
	if isIP(domain) {
		return false
	}
	match, _ := regexp.MatchString("^[a-zA-Z0-9][a-zA-Z0-9-]", domain)
	return match
}

func isIP(ip string) bool {
	return (net.ParseIP(ip) != nil)
}
