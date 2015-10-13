package main

import (
	"bufio"
	"net"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/hoisie/redis"
)

type Hosts struct {
	fileHosts  *FileHosts
	redisHosts *RedisHosts
}

func NewHosts(hs HostsSettings, rs RedisSettings) Hosts {
	fileHosts := &FileHosts{hs.HostsFile, make(map[string]string)}

	var redisHosts *RedisHosts
	if hs.RedisEnable {
		rc := &redis.Client{Addr: rs.Addr(), Db: rs.DB, Password: rs.Password}
		redisHosts = &RedisHosts{rc, hs.RedisKey, make(map[string]string)}
	}

	hosts := Hosts{fileHosts, redisHosts}
	hosts.refresh()
	return hosts

}

/*
1. Match local /etc/hosts file first, remote redis records second
2. Fetch hosts records from /etc/hosts file and redis per minute
*/

func (h *Hosts) Get(domain string, family int) (ip net.IP, ok bool) {

	var sip string

	if sip, ok = h.fileHosts.Get(domain); !ok {
		if h.redisHosts != nil {
			sip, ok = h.redisHosts.Get(domain)
		}
	}

	if sip == "" {
		return nil, false
	}

	switch family {
	case _IP4Query:
		ip = net.ParseIP(sip).To4()
	case _IP6Query:
		ip = net.ParseIP(sip).To16()
	default:
		return nil, false
	}
	return ip, (ip != nil)
}

func (h *Hosts) refresh() {
	ticker := time.NewTicker(time.Minute)
	go func() {
		for {
			h.fileHosts.Refresh()
			if h.redisHosts != nil {
				h.redisHosts.Refresh()
			}
			<-ticker.C
		}
	}()
}

type RedisHosts struct {
	redis *redis.Client
	key   string
	hosts map[string]string
}

func (r *RedisHosts) Get(domain string) (ip string, ok bool) {
	ip, ok = r.hosts[domain]
	return
}

func (r *RedisHosts) Set(domain, ip string) (bool, error) {
	return r.redis.Hset(r.key, domain, []byte(ip))
}

func (r *RedisHosts) Refresh() {
	err := r.redis.Hgetall(r.key, r.hosts)
	if err != nil {
		logger.Warn("Update hosts records from redis failed %s", err)
	} else {
		logger.Debug("Update hosts records from redis")
	}
}

type FileHosts struct {
	file  string
	hosts map[string]string
}

func (f *FileHosts) Get(domain string) (ip string, ok bool) {
	ip, ok = f.hosts[domain]
	return
}

func (f *FileHosts) Refresh() {
	buf, err := os.Open(f.file)
	if err != nil {
		logger.Warn("Update hosts records from file failed %s", err)
		return
	}
	defer buf.Close()

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

		f.hosts[domain] = ip
	}
	logger.Debug("update hosts records from %s", f.file)
}

func (f *FileHosts) isDomain(domain string) bool {
	if f.isIP(domain) {
		return false
	}
	match, _ := regexp.MatchString(`^([a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,6}$`, domain)
	return match
}

func (f *FileHosts) isIP(ip string) bool {
	return (net.ParseIP(ip) != nil)
}
