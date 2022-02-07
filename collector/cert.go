package collector

import (
	"log"
	"sync"

	"git.rabota.space/infrastructure/sbercdn-exporter/apiclient"
	"github.com/prometheus/client_golang/prometheus"
)

type SberCdnCertCollector struct {
	client       *apiclient.SberCdnApiClient
	descriptions map[string]*prometheus.Desc
}

func NewSberCdnCertCollector(client *apiclient.SberCdnApiClient) *SberCdnCertCollector {
	return &SberCdnCertCollector{
		client,
		map[string]*prometheus.Desc{
			"cert_start": prometheus.NewDesc(
				"sbercdn_certificate_valid_since",
				"UNIX time in seconds since EPOCH since which certificate is valid",
				[]string{"account", "cn"},
				nil),
			"cert_end": prometheus.NewDesc(
				"sbercdn_certificate_valid_till",
				"UNIX time in seconds since EPOCH till which certificate is valid",
				[]string{"account", "cn"},
				nil),
		},
	}
}

func (c *SberCdnCertCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, d := range c.descriptions {
		ch <- d
	}
}

func (c *SberCdnCertCollector) Collect(mch chan<- prometheus.Metric) {
	var wag sync.WaitGroup
	getCertMetrics := func(acc string) {
		defer wag.Done()
		cl, err := c.client.GetCertList(acc)
		if err != nil {
			log.Panicln(err)
		}
		certs := cl.Data
		for i := 0; i < len(certs); i++ {
			ci := &certs[i]
			mch <- prometheus.MustNewConstMetric(c.descriptions["cert_start"], prometheus.CounterValue,
				ci.Start, acc, ci.Cn)
			mch <- prometheus.MustNewConstMetric(c.descriptions["cert_end"], prometheus.CounterValue,
				ci.End, acc, ci.Cn)
		}
	}
	for _, acc := range c.client.FindActiveAccounts() {
		wag.Add(1)
		getCertMetrics(acc)
	}
	wag.Wait()
}
