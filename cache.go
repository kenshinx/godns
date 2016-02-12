package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/hoisie/redis"
	"github.com/miekg/dns"
)

type KeyNotFound struct {
	key string
}

func (e KeyNotFound) Error() string {
	return e.key + " " + "not found"
}

type KeyExpired struct {
	Key string
}

func (e KeyExpired) Error() string {
	return e.Key + " " + "expired"
}

type CacheIsFull struct {
}

func (e CacheIsFull) Error() string {
	return "Cache is Full"
}

type SerializerError struct {
}

func (e SerializerError) Error() string {
	return "Serializer error"
}

type Mesg struct {
	Msg    *dns.Msg
	Expire time.Time
}

type Cache interface {
	Get(key string) (Msg *dns.Msg, err error)
	Set(key string, Msg *dns.Msg) error
	Exists(key string) bool
	Remove(key string) error
	Full() bool
}

type Serializer interface {
	Loads([]byte) (*dns.Msg, error)
	Dumps(*dns.Msg) ([]byte, error)
}

type MemoryCache struct {
	Backend  map[string]Mesg
	Expire   time.Duration
	Maxcount int
	mu       sync.RWMutex
}

func (c *MemoryCache) Get(key string) (*dns.Msg, error) {
	c.mu.RLock()
	mesg, ok := c.Backend[key]
	c.mu.RUnlock()
	if !ok {
		return nil, KeyNotFound{key}
	}

	if mesg.Expire.Before(time.Now()) {
		c.Remove(key)
		return nil, KeyExpired{key}
	}

	return mesg.Msg, nil

}

func (c *MemoryCache) Set(key string, msg *dns.Msg) error {
	if c.Full() && !c.Exists(key) {
		return CacheIsFull{}
	}

	expire := time.Now().Add(c.Expire)
	mesg := Mesg{msg, expire}
	c.mu.Lock()
	c.Backend[key] = mesg
	c.mu.Unlock()
	return nil
}

func (c *MemoryCache) Remove(key string) error {
	c.mu.Lock()
	delete(c.Backend, key)
	c.mu.Unlock()
	return nil
}

func (c *MemoryCache) Exists(key string) bool {
	c.mu.RLock()
	_, ok := c.Backend[key]
	c.mu.RUnlock()
	return ok
}

func (c *MemoryCache) Length() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.Backend)
}

func (c *MemoryCache) Full() bool {
	// if Maxcount is zero. the cache will never be full.
	if c.Maxcount == 0 {
		return false
	}
	return c.Length() >= c.Maxcount
}

/*
Memcached backend
*/

func NewMemcachedCache(servers []string, expire int32) *MemcachedCache {
	c := memcache.New(servers...)
	return &MemcachedCache{
		backend:    c,
		serializer: &JsonSerializer{},
		expire:     expire,
	}
}

type MemcachedCache struct {
	backend    *memcache.Client
	serializer Serializer
	expire     int32
}

func (m *MemcachedCache) Set(key string, msg *dns.Msg) error {
	var val []byte
	var err error

	// handle cases for negacache where it sets nil values
	if msg == nil {
		val = []byte("nil")
	} else {
		val, err = m.serializer.Dumps(msg)
	}
	if err != nil {
		return err
	}
	return m.backend.Set(&memcache.Item{Key: key, Value: val, Expiration: m.expire})
}

func (m *MemcachedCache) Get(key string) (msg *dns.Msg, err error) {
	item, err := m.backend.Get(key)
	if err != nil {
		err = KeyNotFound{key}
		return
	}
	msg, err = m.serializer.Loads(item.Value)
	return
}

func (m *MemcachedCache) Exists(key string) bool {
	_, err := m.backend.Get(key)
	if err != nil {
		return true
	}
	return false
}

func (m *MemcachedCache) Remove(key string) error {
	return m.backend.Delete(key)
}

func (m *MemcachedCache) Full() bool {
	// memcache is never full (LRU)
	return false
}

/*
TODO: Redis cache Backend
*/

type RedisCache struct {
	Backend    *redis.Client
	Serializer JsonSerializer
	Expire     time.Duration
	Maxcount   int
}

func (c *RedisCache) Get() {

}

func (c *RedisCache) Set() {

}

func (c *RedisCache) Remove() {

}

func KeyGen(q Question) string {
	h := md5.New()
	h.Write([]byte(q.String()))
	x := h.Sum(nil)
	key := fmt.Sprintf("%x", x)
	return key
}

/* we need to define marsheling to encode and decode
 */
type JsonSerializer struct {
}

func (*JsonSerializer) Dumps(mesg *dns.Msg) (encoded []byte, err error) {
	encoded, err = json.Marshal(*mesg)
	return
}

func (*JsonSerializer) Loads(data []byte) (*dns.Msg, error) {
	var mesg dns.Msg
	err := json.Unmarshal(data, &mesg)
	return &mesg, err
}
