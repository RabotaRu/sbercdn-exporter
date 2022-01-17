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
	ApiUrl       string            `yaml:"api_url"`
	URNs         map[string]string `yaml:"URNs"`
	Auth         Auth              `yaml:"auth"`
	MaxQueryTime time.Duration     `yaml:"max_query_time"`
}

type ListenConf struct {
	Address     string `yaml:"address"`
	CertFile    string `yaml:"cert_file"`
	PrivkeyFile string `yaml:"privkey_file"`
}

type AppConf struct {
	Listen ListenConf `yaml:"listen"`
	Client ClientConf `yaml:"API"`
}
