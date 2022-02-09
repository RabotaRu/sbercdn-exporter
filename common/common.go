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
	ScrapeTimeOffset time.Duration `yaml:"scrape_time_offset"`
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
		if fval.Kind() == reflect.Struct {
			ReadConfigFromEnv(var_name, tag, fval.Addr().Interface())
		}
		var var_val string
		var ok bool
		if var_val, ok = os.LookupEnv(var_name); !ok {
			continue
		}
		switch fval.Kind() { //nolint: exhaustive
		case reflect.String:
			fval.SetString(var_val)
		case reflect.Int64:
			// TODO: write more safe value parsing
			if d, err := time.ParseDuration(var_val); err != nil {
				log.Panicf("Failed to parse duration from env var %v with value %v: %v\n",
					var_name, var_val, err)
			} else {
				fval.SetInt(d.Nanoseconds())
			}
		case reflect.Slice:
			tmpslice := strings.Split(var_val, ",")
			var slice []string
			for i := range tmpslice {
				v := strings.TrimSpace(tmpslice[i])
				if v != "" {
					slice = append(slice, v)
				}
			}
			fval.Set(reflect.ValueOf(slice))
		default:
			log.Println("Found not catched env var", var_name, "with value", var_val, "and type", fval.Kind())
		}
	}
}
