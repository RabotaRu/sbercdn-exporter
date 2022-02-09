package collector

import (
	"context"
	"log"
	"sync"

	"git.rabota.space/infrastructure/sbercdn-exporter/apiclient"
	cmn "git.rabota.space/infrastructure/sbercdn-exporter/common"
	"github.com/prometheus/client_golang/prometheus"
)

type SberCdnCertCollector struct {
	client       *apiclient.SberCdnApiClient
	descriptions map[string]*prometheus.Desc
	api_group    string
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
		"certList",
	}
}

func (c *SberCdnCertCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, d := range c.descriptions {
		ch <- d
	}
}

func (c *SberCdnCertCollector) Collect(mch chan<- prometheus.Metric) {
	var wag sync.WaitGroup
	ctx_root, ctx_cancel := context.WithTimeout(context.WithValue(context.Background(), cmn.CtxKey("api_group"), c.api_group), c.client.MaxQueryTime)
	defer ctx_cancel()

	getCertMetrics := func(ctx context.Context) {
		defer wag.Done()
		defer func() {
			if r := recover(); r != nil {
				log.Println(r)
			}
		}()
		var ok bool
		var account string
		if account, ok = ctx.Value(cmn.CtxKey("account")).(string); !ok {
			log.Panicln("Oops account is not a string or is empty")
		}
		cl, err := c.client.GetCertList(ctx)
		if err != nil {
			log.Panicln(err)
		}
		certs := cl.Data
		for i := 0; i < len(certs); i++ {
			ci := &certs[i]
			mch <- prometheus.MustNewConstMetric(c.descriptions["cert_start"], prometheus.CounterValue,
				ci.Start, account, ci.Cn)
			mch <- prometheus.MustNewConstMetric(c.descriptions["cert_end"], prometheus.CounterValue,
				ci.End, account, ci.Cn)
		}
	}
	for _, acc := range c.client.FindActiveAccounts(ctx_root) {
		wag.Add(1)
		getCertMetrics(context.WithValue(ctx_root, cmn.CtxKey("account"), acc))
	}
	wag.Wait()
}
