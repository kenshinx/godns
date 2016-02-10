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
Match local /etc/hosts file first, remote redis records second
*/
func (h *Hosts) Get(domain string, family int) ([]net.IP, bool) {

	var sips []string
	var ip net.IP
	var ips []net.IP

	sips, ok := h.fileHosts.Get(strings.ToLower(domain))
	if !ok {
		if h.redisHosts != nil {
			sips, ok = h.redisHosts.Get(strings.ToLower(domain))
		}
	}

	if sips == nil {
		return nil, false
	}

	for _, sip := range sips {
		switch family {
		case _IP4Query:
			ip = net.ParseIP(sip).To4()
		case _IP6Query:
			ip = net.ParseIP(sip).To16()
		default:
			continue
		}
		if ip != nil {
			ips = append(ips, ip)
		}
	}

	return ips, (ips != nil)
}

/*
Update hosts records from /etc/hosts file and redis per minute
*/
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

type BaseHosts struct {
	hosts map[string]string
}

type RedisHosts struct {
	redis *redis.Client
	key   string
	hosts map[string]string
}

func (r *RedisHosts) Get(domain string) ([]string, bool) {
	ip, ok := r.hosts[strings.ToLower(domain)]
	if ok {
		return strings.Split(ip, ","), true
	}

	for host, ip := range r.hosts {
		if strings.HasPrefix(host, "*.") {
			upperLevelDomain := strings.Split(host, "*.")[1]
			if strings.HasSuffix(domain, upperLevelDomain) {
				return strings.Split(ip, ","), true
			}
		}
	}
	return nil, false
}

func (r *RedisHosts) Set(domain, ip string) (bool, error) {
	return r.redis.Hset(r.key, strings.ToLower(domain), []byte(ip))
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

func (f *FileHosts) Get(domain string) ([]string, bool) {
	ip, ok := f.hosts[strings.ToLower(domain)]
	if !ok {
		return nil, false
	}
	return []string{ip}, true
}

func (f *FileHosts) Refresh() {
	buf, err := os.Open(f.file)
	if err != nil {
		logger.Warn("Update hosts records from file failed %s", err)
		return
	}
	defer buf.Close()

	f.clear()

	scanner := bufio.NewScanner(buf)
	for scanner.Scan() {

		line := scanner.Text()
		line = strings.TrimSpace(strings.ToLower(line))

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

		f.hosts[strings.ToLower(domain)] = ip
	}
	logger.Debug("update hosts records from %s", f.file)
}

func (f *FileHosts) clear() {
	f.hosts = make(map[string]string)
}

func (f *FileHosts) isDomain(domain string) bool {
	if f.isIP(domain) {
		return false
	}
	match, _ := regexp.MatchString(`^([a-zA-Z0-9\*]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,6}$`, domain)
	return match
}

func (f *FileHosts) isIP(ip string) bool {
	return (net.ParseIP(ip) != nil)
}
