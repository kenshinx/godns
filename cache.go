package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/miekg/dns"
	"time"
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

type Cache interface {
	Get(string) ([]byte, error)
	Set(string, *dns.Msg) error
	Exists(string) bool
	Remove()
	Length() int
}

type MemoryCache struct {
	backend    map[string]string
	serializer *JsonSerializer
	expire     time.Duration
	maxcount   int
}

func (c *MemoryCache) Get(key string) ([]byte, error) {
	fmt.Println(c.backend)
	data, ok := c.backend[key]
	if !ok {
		return nil, KeyNotFound{key}
	}
	return []byte(data), nil

	// mesg := new(dns.Msg)
	// if err := c.serializer.Loads([]byte(data), &mesg); err != nil {
	// 	fmt.Println(err)
	// 	return nil, SerializerError{}
	// }
	// return mesg, nil

}

func (c *MemoryCache) Set(key string, mesg *dns.Msg) error {
	if c.Full() && !c.Exists(key) {
		return CacheIsFull{}
	}
	// data, err := c.serializer.Dumps(mesg)

	// if err != nil {
	// 	return SerializerError{}
	// }

	c.backend[key] = mesg.String()
	return nil
}

func (c *MemoryCache) Remove() {

}

func (c *MemoryCache) Exists(key string) bool {
	_, ok := c.backend[key]
	return ok
}

func (c *MemoryCache) Length() int {
	return len(c.backend)
}

func (c *MemoryCache) Full() bool {
	// if maxcount is zero. the cache will never be full.
	if c.maxcount == 0 {
		return false
	}
	return c.Length() >= c.maxcount
}

// type RedisCache struct{
// backend redis.client
// }

// func (c *RedisCache) Get(key string) {

// }

// func (c *RedisCache) Set() {

// }

// func (c &RedisCache) Remove(){

// }

func KeyGen(q Question) string {
	h := md5.New()
	h.Write([]byte(q.String()))
	x := h.Sum(nil)
	key := fmt.Sprintf("%x", x)
	return key
}

type JsonSerializer struct {
}

func (*JsonSerializer) Dumps(mesg *dns.Msg) (encoded []byte, err error) {
	encoded, err = json.Marshal(mesg)
	return
}

func (*JsonSerializer) Loads(data []byte, mesg **dns.Msg) error {
	err := json.Unmarshal(data, mesg)
	return err

}
