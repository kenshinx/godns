package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hoisie/redis"
)

const AUDIT_LOG_OUTPUT_BUFFER = 1024

type AuditLogger interface {
	Run()
	Write(mesg *AuditMesg)
}

type AuditMesg struct {
	RemoteAddr string    `json:"remoteaddr"`
	Domain     string    `json:"domain"`
	QType      string    `json:"qtype"`
	Timestamp  time.Time `json:"timestamp"`
}

func NewAuditMessage(remoteAddr string, domain string, qtype string) *AuditMesg {
	return &AuditMesg{
		RemoteAddr: remoteAddr,
		Domain:     domain,
		QType:      qtype,
		Timestamp:  time.Now(),
	}
}

type RedisAuditLogger struct {
	backend *redis.Client
	mesgs   chan *AuditMesg
	expire  int64
}

func NewRedisAuditLogger(rs RedisSettings, expire int64) AuditLogger {
	rc := &redis.Client{Addr: rs.Addr(), Db: rs.DB, Password: rs.Password}
	auditLogger := &RedisAuditLogger{
		backend: rc,
		mesgs:   make(chan *AuditMesg, AUDIT_LOG_OUTPUT_BUFFER),
		expire:  expire,
	}
	go auditLogger.Run()
	return auditLogger
}

func (rl *RedisAuditLogger) Run() {
	for {
		select {
		case mesg := <-rl.mesgs:
			jsonMesg, err := json.Marshal(mesg)
			if err != nil {
				logger.Error("Can't write to redis audit log: %v", err)
				continue
			}
			redisKey := fmt.Sprintf("audit-%s:00", mesg.Timestamp.Format("2006-01-02T15"))
			err = rl.backend.Rpush(redisKey, jsonMesg)
			if err != nil {
				logger.Error("Can't write to redis audit log: %v", err)
				continue
			}
			_, err = rl.backend.Expire(redisKey, rl.expire)
			if err != nil {
				logger.Error("Can't set expiration for redis audit log: %v", err)
				continue
			}
		}
	}
}

func (rl *RedisAuditLogger) Write(mesg *AuditMesg) {
	rl.mesgs <- mesg
}
