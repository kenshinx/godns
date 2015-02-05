package main

import (
	"flag"
	"fmt"
	"github.com/BurntSushi/toml"
	"os"
	"strconv"
)

var (
	settings Settings
)

type Settings struct {
	Version      string
	Debug        bool
	Server       DNSServerSettings `toml:"server"`
	ResolvConfig ResolvSettings    `toml:"resolv"`
	Redis        RedisSettings     `toml:"redis"`
	Log          LogSettings       `toml:"log"`
	Cache        CacheSettings     `toml:"cache"`
	Hosts        HostsSettings     `toml:"hosts"`
}

type ResolvSettings struct {
	ResolvFile string `toml:"resolv-file"`
	Timeout    int
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

func (s RedisSettings) Addr() string {
	return s.Host + ":" + strconv.Itoa(s.Port)
}

type LogSettings struct {
	File string
}

type CacheSettings struct {
	Backend  string
	Expire   int
	Maxcount int
}

type HostsSettings struct {
	Enable      bool
	HostsFile   string `toml:"host-file"`
	RedisEnable bool   `toml:"redis-enable"`
	RedisKey    string `toml:"redis-key"`
	TTL         uint32 `toml:"ttl"`
}

func init() {

	var configFile string

	flag.StringVar(&configFile, "c", "godns.conf", "Look for godns toml-formatting config file in this directory")
	flag.Parse()

	if _, err := toml.DecodeFile(configFile, &settings); err != nil {
		fmt.Printf("%s is not a valid toml config file\n", configFile)
		fmt.Println(err)
		os.Exit(1)
	}

}
