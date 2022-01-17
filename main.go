package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"strings"
	"syscall"
	"time"

	"git.rabota.space/infrastructure/sbercdn-exporter/api_client"

	// nolint: stylecheck
	. "git.rabota.space/infrastructure/sbercdn-exporter/common"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/yaml.v2"
)

var (
	Version = "dev"
	config  = AppConf{
		Client: ClientConf{
			ApiUrl: "https://api.cdn.sber.cloud",
			Auth: Auth{
				Urn:           "/app/oauth/v1/token/",
				TokenLifetime: time.Hour * 6,
			},
			URNs:         map[string]string{"CertList": "/app/ssl/v1/account/{{ auth.id }}/certificate/"},
			MaxQueryTime: time.Second * 10,
		},
		Listen: ListenConf{Address: ":9921"},
	}
)

func readConfigFromEnv(prefix, tag string, c interface{}) {
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
			readConfigFromEnv(var_name, tag, fval.Addr().Interface())
		default:
			if v, ok := os.LookupEnv(var_name); ok {
				log.Println("Found not catched env var", var_name, "with value", v, "and type", fval.Kind())
			}
		}
	}
}

func readConfigFromFile(cf string, config *AppConf) (err error) {
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

func main() {
	flags := make(map[string]bool)
	configPath := flag.String("config", "sbercdn-exporter.yaml", "Path to config file in YAML format")
	showVersion := flag.Bool("version", false, "Show version")
	flag.Parse()
	if *showVersion {
		log.Fatalln(Version)
	}
	flag.Visit(func(f *flag.Flag) { flags[f.Name] = true })

	err := readConfigFromFile(*configPath, &config)
	if err != nil && (flags["config"] || !os.IsNotExist(err)) {
		log.Fatalln("could not read config file:", err)
	}
	readConfigFromEnv("SC", "yaml", &config)
	// log.Printf("config is: %+v\n", config)
	if strings.HasPrefix(config.Listen.Address, ":") {
		config.Listen.Address = "0.0.0.0" + config.Listen.Address
	}

	apiClient := api_client.NewSberCdnApiClient(&config.Client)
	collector := NewSberCdnCollector(apiClient)
	prometheus.MustRegister(collector)

	http.Handle("/metrics", promhttp.Handler())

	srv := http.Server{
		Addr: config.Listen.Address,
	}
	idleConnsClosed := make(chan struct{})
	go func() {
		siginterm := make(chan os.Signal, 1)
		signal.Notify(siginterm, syscall.SIGINT, syscall.SIGTERM)
		<-siginterm

		// We received an interrupt signal, shut down.
		if err := srv.Shutdown(context.Background()); err != nil {
			// Error from closing listeners, or context timeout:
			log.Printf("HTTP server Shutdown: %v", err)
		}
		close(idleConnsClosed)
	}()

	if config.Listen.CertFile != "" && config.Listen.PrivkeyFile != "" {
		log.Printf("Begin listening on https://%v/auth", config.Listen.Address)
		err = srv.ListenAndServeTLS(config.Listen.CertFile, config.Listen.PrivkeyFile)
	} else {
		log.Printf("Begin listening on http://%v/auth", config.Listen.Address)
		err = srv.ListenAndServe()
	}
	if !errors.Is(http.ErrServerClosed, err) {
		// Error starting or closing listener:
		log.Fatalf("HTTP server ListenAndServe: %v", err)
	}

	<-idleConnsClosed
}
