package main

import (
	"flag"
	"fmt"
	"github.com/BurntSushi/toml"
	"os"
	"time"
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

type LogSettings struct {
	File string
}

type CacheSettings struct {
	Backend  string
	Expire   time.Duration
	Maxcount uint32
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
