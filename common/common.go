package common

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

type ClientConf struct {
	Url              string        `yaml:"url"`
	AuthUrn          string        `yaml:"auth_urn"`
	Username         string        `yaml:"username"`
	Password         string        `yaml:"password"`
	Accounts         []string      `yaml:"accounts"`
	TokenLifetime    time.Duration `yaml:"token_lifetime"`
	MaxQueryTime     time.Duration `yaml:"max_query_time"`
	ScrapeInterval   time.Duration `yaml:"scrape_interval"`
	ScrapeTimeOffset time.Duration `yaml:"scrape_interval"`
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

func ReadConfigFromFile(cf string, config *AppConf) (err error) {
	buf := make([]byte, 4096)

	file, err := os.Open(cf)
	if err != nil {
		return
	}
	defer file.Close()

	n, err := file.Read(buf)
	if err != nil {
		return
	}

	err = yaml.Unmarshal(buf[:n], &config)
	if err != nil {
		return fmt.Errorf("cannot unmarshal data: %w", err)
	}

	return
}

func ReadConfigFromEnv(prefix, tag string, c interface{}) {
	cv := reflect.ValueOf(c)
	ce := cv.Elem()
	ct := reflect.Indirect(cv).Type()
	for i := 0; i < ct.NumField(); i++ {
		ft := ct.Field(i)
		fval := ce.Field(i)
		var_name := strings.ReplaceAll(strings.Trim(strings.ToUpper(
			prefix+" "+ft.Tag.Get(tag)),
			" \t\n"),
			" ", "_")
		switch fval.Kind() { //nolint: exhaustive
		case reflect.String:
			if v, ok := os.LookupEnv(var_name); ok {
				fval.SetString(v)
			}
		case reflect.Int64:
			// TODO: write more safe value parsing
			if v, ok := os.LookupEnv(var_name); ok {
				if v, err := time.ParseDuration(v); err != nil {
					log.Println("Found env var", var_name, "with value", v)
				} else {
					fval.SetInt(v.Nanoseconds())
				}
			}
		case reflect.Struct:
			ReadConfigFromEnv(var_name, tag, fval.Addr().Interface())
		default:
			if v, ok := os.LookupEnv(var_name); ok {
				log.Println("Found not catched env var", var_name, "with value", v, "and type", fval.Kind())
			}
		}
	}
}
