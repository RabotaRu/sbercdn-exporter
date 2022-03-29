package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/RabotaRu/sbercdn-exporter/apiclient"
	col "github.com/RabotaRu/sbercdn-exporter/collector"
	cmn "github.com/RabotaRu/sbercdn-exporter/common"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	Version = "dev"
	config  = cmn.AppConf{
		Client: cmn.ClientConf{
			Url:              "https://api.cdn.sber.cloud",
			AuthUrn:          "/app/oauth/v1/token/",
			TokenLifetime:    6 * time.Hour,
			MaxQueryTime:     10 * time.Second,
			ScrapeInterval:   time.Minute,
			ScrapeTimeOffset: 5 * time.Minute,
		},
		Listen: cmn.ListenConf{Address: ":9921"},
	}
)

func main() {
	flags := make(map[string]bool)
	configPath := flag.String("config", "sbercdn-exporter.yaml", "Path to config file in YAML format")
	showVersion := flag.Bool("version", false, "Show version")
	flag.Parse()
	if *showVersion {
		log.Fatalln(Version)
	}
	flag.Visit(func(f *flag.Flag) { flags[f.Name] = true })

	err := cmn.ReadConfigFromFile(*configPath, &config)
	if err != nil && (flags["config"] || !os.IsNotExist(err)) {
		log.Fatalln("could not read config file:", err)
	}
	cmn.ReadConfigFromEnv("SCE", "yaml", &config)
	if strings.HasPrefix(config.Listen.Address, ":") {
		config.Listen.Address = "0.0.0.0" + config.Listen.Address
	}

	if config.Client.Username == "" || config.Client.Username == "" {
		log.Fatalln("API username or password is missing or empty")
	}

	apiClient, err := apiclient.NewSberCdnApiClient(&config.Client)
	if err != nil {
		log.Fatalf("failed to start api client: %v", err)
	}

	prometheus.MustRegister(col.NewSberCdnCertCollector(apiClient))
	prometheus.MustRegister(col.NewSberCdnStatsCollector(apiClient))

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
		log.Printf("Begin listening on https://%v/metircs", config.Listen.Address)
		err = srv.ListenAndServeTLS(config.Listen.CertFile, config.Listen.PrivkeyFile)
	} else {
		log.Printf("Begin listening on http://%v/metrics", config.Listen.Address)
		err = srv.ListenAndServe()
	}
	if !errors.Is(http.ErrServerClosed, err) {
		// Error starting or closing listener:
		log.Fatalf("HTTP server ListenAndServe: %v", err)
	}

	<-idleConnsClosed
}
