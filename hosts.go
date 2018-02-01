package main

import (
	"bufio"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/hoisie/redis"
	"golang.org/x/net/publicsuffix"
)

type Hosts struct {
	fileHosts       *FileHosts
	redisHosts      *RedisHosts
	refreshInterval time.Duration
}

func NewHosts(hs HostsSettings, rs RedisSettings) Hosts {
	fileHosts := &FileHosts{
		file:  hs.HostsFile,
		hosts: make(map[string]string),
	}

	var redisHosts *RedisHosts
	if hs.RedisEnable {
		rc := &redis.Client{Addr: rs.Addr(), Db: rs.DB, Password: rs.Password}
		redisHosts = &RedisHosts{
			redis: rc,
			key:   hs.RedisKey,
			hosts: make(map[string]string),
		}
	}

	hosts := Hosts{fileHosts, redisHosts, time.Second * time.Duration(hs.RefreshInterval)}
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

	sips, ok := h.fileHosts.Get(domain)
	if !ok {
		if h.redisHosts != nil {
			sips, ok = h.redisHosts.Get(domain)
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
	ticker := time.NewTicker(h.refreshInterval)
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
	mu    sync.RWMutex
}

func (r *RedisHosts) Get(domain string) ([]string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	domain = strings.ToLower(domain)
	ip, ok := r.hosts[domain]
	if ok {
		return strings.Split(ip, ","), true
	}

	sld, err := publicsuffix.EffectiveTLDPlusOne(domain)
	if err != nil {
		return nil, false
	}

	for host, ip := range r.hosts {
		if strings.HasPrefix(host, "*.") {
			old, err := publicsuffix.EffectiveTLDPlusOne(host)
			if err != nil {
				continue
			}
			if sld == old {
				return strings.Split(ip, ","), true
			}
		}
	}
	return nil, false
}

func (r *RedisHosts) Set(domain, ip string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.redis.Hset(r.key, strings.ToLower(domain), []byte(ip))
}

func (r *RedisHosts) Refresh() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.clear()
	err := r.redis.Hgetall(r.key, r.hosts)
	if err != nil {
		logger.Warn("Update hosts records from redis failed %s", err)
	} else {
		logger.Debug("Update hosts records from redis")
	}
}

func (r *RedisHosts) clear() {
	r.hosts = make(map[string]string)
}

type FileHosts struct {
	file  string
	hosts map[string]string
	mu    sync.RWMutex
}

func (f *FileHosts) Get(domain string) ([]string, bool) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	domain = strings.ToLower(domain)
	ip, ok := f.hosts[domain]
	if ok {
		return []string{ip}, true
	}

	sld, err := publicsuffix.EffectiveTLDPlusOne(domain)
	if err != nil {
		return nil, false
	}

	for host, ip := range f.hosts {
		if strings.HasPrefix(host, "*.") {
			old, err := publicsuffix.EffectiveTLDPlusOne(host)
			if err != nil {
				continue
			}
			if sld == old {
				return []string{ip}, true
			}
		}
	}

	return nil, false
}

func (f *FileHosts) Refresh() {
	buf, err := os.Open(f.file)
	if err != nil {
		logger.Warn("Update hosts records from file failed %s", err)
		return
	}
	defer buf.Close()

	f.mu.Lock()
	defer f.mu.Unlock()

	f.clear()

	scanner := bufio.NewScanner(buf)
	for scanner.Scan() {

		line := scanner.Text()
		line = strings.TrimSpace(line)
		line = strings.Replace(line, "\t", " ", -1)

		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		sli := strings.Split(line, " ")

		if len(sli) < 2 {
			continue
		}

		ip := sli[0]
		if !isIP(ip) {
			continue
		}

		// Would have multiple columns of domain in line.
		// Such as "127.0.0.1  localhost localhost.domain" on linux.
		// The domains may not strict standard, like "local" so don't check with f.isDomain(domain).
		for i := 1; i <= len(sli)-1; i++ {
			domain := strings.TrimSpace(sli[i])
			if domain == "" {
				continue
			}

			f.hosts[strings.ToLower(domain)] = ip
		}
	}
	logger.Debug("update hosts records from %s, total %d records.", f.file, len(f.hosts))
}

func (f *FileHosts) clear() {
	f.hosts = make(map[string]string)
}
