package main

import (
	"bufio"
	"github.com/hoisie/redis"
	"os"
	"regexp"
	"strings"
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

func (h *Hosts) Get(domain string) (ip string, ok bool) {
	if ip, ok = h.FileHosts[domain]; ok {
		return
	}
	if ip, ok = h.RedisHosts.Get(domain); ok {
		return
	}
	return "", false
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
		if !f.isDomain(domain) || !f.isIP(ip) {
			continue
		}

		hosts[domain] = ip
	}
	return hosts
}

func (f *FileHosts) isDomain(domain string) bool {
	match, _ := regexp.MatchString("^[a-z]", domain)
	return match
}

func (f *FileHosts) isIP(ip string) bool {
	match, _ := regexp.MatchString("^[1-9]", ip)
	return match
}
