package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hoisie/redis"
	_ "github.com/lib/pq"
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

type PostgresqlAuditLogger struct {
	backend *sql.DB
	mesgs   chan *AuditMesg
	expire  int64
}

func NewPostgresqlAuditLogger(ps PostgresqlSettings, expire int64) AuditLogger {
	connStr := fmt.Sprintf(`
                host=%s port=%d
                user=%s password=%s
                dbname=%s sslmode=%s
                sslcert=%s sslkey=%s
                sslrootcert=%s
                `,
		ps.Host, ps.Port,
		ps.User, ps.Password,
		ps.DB, ps.Sslmode,
		ps.Sslcert, ps.Sslkey,
		ps.Sslrootcert,
	)
	pc, err := sql.Open("postgres", connStr)
	if err != nil {
		logger.Error("Can't connect to audit log postgresql: %v", err)
	}
	rows, err := pc.Query(`
                CREATE TABLE IF NOT EXISTS audit (
                        id BIGSERIAL NOT NULL,
                        remoteaddr TEXT,
                        domain TEXT,
                        qtype TEXT,
                        timestamp TIMESTAMP
                )
        `)
	defer rows.Close()
	auditLogger := &PostgresqlAuditLogger{
		backend: pc,
		mesgs:   make(chan *AuditMesg, AUDIT_LOG_OUTPUT_BUFFER),
		expire:  expire,
	}
	go auditLogger.Run()
	go auditLogger.Expire()
	return auditLogger
}

func (pl *PostgresqlAuditLogger) Run() {
	for {
		select {
		case mesg := <-pl.mesgs:
			rows, err := pl.backend.Query(`INSERT INTO audit (remoteaddr, domain, qtype, timestamp) VALUES ($1, $2, $3, $4)`,
				mesg.RemoteAddr, mesg.Domain, mesg.QType, mesg.Timestamp,
			)
			rows.Close()
			if err != nil {
				logger.Error("Can't write to postgresql audit log: %v", err)
				continue
			}
		}
	}
}

func (pl *PostgresqlAuditLogger) Write(mesg *AuditMesg) {
	pl.mesgs <- mesg
}

func (pl *PostgresqlAuditLogger) Expire() {
	for {
		expireTime := time.Now().Add(time.Duration(-pl.expire) * time.Second)
		rows, err := pl.backend.Query(`DELETE FROM audit WHERE timestamp < $1`, expireTime)
		rows.Close()
		if err != nil {
			logger.Error("Can't expire postgresql audit log: %v", err)
		}
		time.Sleep(time.Duration(pl.expire) * time.Second / 2)
	}
}
