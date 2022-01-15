package main

import (
	"log"

	"github.com/prometheus/client_golang/prometheus"
)

type SberCdnCollector struct {
	client       *SberCdnApiClient
	descriptions map[string]*prometheus.Desc
}

func NewSberCdnCollector(client *SberCdnApiClient) *SberCdnCollector {
	return &SberCdnCollector{
		client,
		map[string]*prometheus.Desc{
			"cert_start": prometheus.NewDesc(
				"certificate_valid_since",
				"Float64 representing UNIX time in seconds since EPOCH since which certificate is valid",
				[]string{"alt", "cn", "comment", "issuer"},
				nil),
			"cert_end": prometheus.NewDesc(
				"certificate_valid_till",
				"Float64 representing UNIX time in seconds since EPOCH till which certificate is valid",
				[]string{"alt", "cn", "comment", "issuer"},
				nil),
		},
	}
}

func (c *SberCdnCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, d := range c.descriptions {
		ch <- d
	}
}

func (c *SberCdnCollector) Collect(mch chan<- prometheus.Metric) {
	cl, err := c.client.GetCertList()
	if err != nil {
		log.Panicln(err)
	}
	for i := 0; i < len(cl); i++ {
		ci := &cl[i]
		mch <- prometheus.MustNewConstMetric(c.descriptions["cert_start"], prometheus.CounterValue, ci.Start,
			ci.Alt, ci.Cn, ci.Comment, ci.Issuer)
		mch <- prometheus.MustNewConstMetric(c.descriptions["cert_end"], prometheus.CounterValue,
			ci.End, ci.Alt, ci.Cn, ci.Comment, ci.Issuer)
	}
}
