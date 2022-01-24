package common

import (
	"time"
)

type Auth struct {
	Id            string        `yaml:"id"`
	Username      string        `yaml:"username"`
	Password      string        `yaml:"password"`
	Urn           string        `yaml:"urn"`
	TokenLifetime time.Duration `yaml:"token_lifetime"`
}

type ClientConf struct {
	Url            string        `yaml:"url"`
	Auth           Auth          `yaml:"auth"`
	MaxQueryTime   time.Duration `yaml:"max_query_time"`
	ScrapeInterval time.Duration `yaml:"scrape_interval"`
}

type ListenConf struct {
	Address     string `yaml:"address"`
	CertFile    string `yaml:"cert_file"`
	PrivkeyFile string `yaml:"privkey_file"`
}

type AppConf struct {
	Listen ListenConf `yaml:"listen"`
	Client ClientConf `yaml:"api"`
}
