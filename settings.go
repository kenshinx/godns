package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/BurntSushi/toml"
)

var (
	settings Settings
)

var LogLevelMap = map[string]int{
	"DEBUG":  LevelDebug,
	"INFO":   LevelInfo,
	"NOTICE": LevelNotice,
	"WARN":   LevelWarn,
	"ERROR":  LevelError,
}

type Settings struct {
	Version      string
	Debug        bool
	Server       DNSServerSettings `toml:"server"`
	ResolvConfig ResolvSettings    `toml:"resolv"`
	Redis        RedisSettings     `toml:"redis"`
	Memcache     MemcacheSettings  `toml:"memcache"`
	Log          LogSettings       `toml:"log"`
	Cache        CacheSettings     `toml:"cache"`
	Hosts        HostsSettings     `toml:"hosts"`
}

type ResolvSettings struct {
	Timeout        int
	Interval       int
	SetEDNS0       bool
	ServerListFile string `toml:"server-list-file"`
	ResolvFile     string `toml:"resolv-file"`
}

type DNSServerSettings struct {
	Host string
	Port int
}

type RedisSettings struct {
	Host     string
	Port     int
	DB       int
	Password string
}

type MemcacheSettings struct {
	Servers []string
}

func (s RedisSettings) Addr() string {
	return s.Host + ":" + strconv.Itoa(s.Port)
}

type LogSettings struct {
	Stdout bool
	File   string
	Level  string
}

func (ls LogSettings) LogLevel() int {
	l, ok := LogLevelMap[ls.Level]
	if !ok {
		panic("Config error: invalid log level: " + ls.Level)
	}
	return l
}

type CacheSettings struct {
	Backend  string
	Expire   int
	Maxcount int
}

type HostsSettings struct {
	Enable          bool
	HostsFile       string `toml:"host-file"`
	RedisEnable     bool   `toml:"redis-enable"`
	RedisKey        string `toml:"redis-key"`
	TTL             uint32 `toml:"ttl"`
	RefreshInterval uint32 `toml:"refresh-interval"`
}

func init() {

	var configFile string

	flag.StringVar(&configFile, "c", "./etc/godns.conf", "Look for godns toml-formatting config file in this directory")
	flag.Parse()

	if _, err := toml.DecodeFile(configFile, &settings); err != nil {
		fmt.Printf("%s is not a valid toml config file\n", configFile)
		fmt.Println(err)
		os.Exit(1)
	}

}
