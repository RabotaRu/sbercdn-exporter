package collector

import (
	"log"

	"git.rabota.space/infrastructure/sbercdn-exporter/api_client"
	"github.com/prometheus/client_golang/prometheus"
)

type SberCdnCertCollector struct {
	client       *api_client.SberCdnApiClient
	descriptions map[string]*prometheus.Desc
}

func NewSberCdnCertCollector(client *api_client.SberCdnApiClient) *SberCdnCertCollector {
	return &SberCdnCertCollector{
		client,
		map[string]*prometheus.Desc{
			"cert_start": prometheus.NewDesc(
				"sbercdn_certificate_valid_since",
				"UNIX time in seconds since EPOCH since which certificate is valid",
				[]string{"cn"},
				nil),
			"cert_end": prometheus.NewDesc(
				"sbercdn_certificate_valid_till",
				" UNIX time in seconds since EPOCH till which certificate is valid",
				[]string{"cn"},
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
	cl, err := c.client.GetCertList()
	if err != nil {
		log.Panicln(err)
	}
	certs := cl.Data
	for i := 0; i < len(certs); i++ {
		ci := &certs[i]
		mch <- prometheus.MustNewConstMetric(c.descriptions["cert_start"], prometheus.CounterValue,
			ci.Start, ci.Cn)
		mch <- prometheus.MustNewConstMetric(c.descriptions["cert_end"], prometheus.CounterValue,
			ci.End, ci.Cn)
	}
}
