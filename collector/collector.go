package collector

import (
	"log"
	"time"

	"git.rabota.space/infrastructure/sbercdn-exporter/api_client"
	"github.com/prometheus/client_golang/prometheus"
)

type Metric struct {
	desc      *prometheus.Desc
	valueType prometheus.ValueType
}

type SberCdnSummaryCollector struct {
	client  *api_client.SberCdnApiClient
	metrics map[string]*Metric
}

func NewSberCdnSummaryCollector(client *api_client.SberCdnApiClient) *SberCdnSummaryCollector {
	return &SberCdnSummaryCollector{
		client,
		map[string]*Metric{
			"sum_bw": {prometheus.NewDesc(
				"sbercdn_summary_bandwidth",
				"Float64 representing UNIX time in seconds since EPOCH since which certificate is valid",
				[]string{},
				nil), prometheus.GaugeValue},
			"sum_ratio": {prometheus.NewDesc(
				"sbercdn_summary_cache_ration",
				"Float64 representing UNIX time in seconds since EPOCH since which certificate is valid",
				[]string{},
				nil), prometheus.GaugeValue},
			"sum_hits": {prometheus.NewDesc(
				"sbercdn_summary_hits",
				"Float64 representing UNIX time in seconds since EPOCH since which certificate is valid",
				[]string{},
				nil), prometheus.GaugeValue},
			"sum_traf": {prometheus.NewDesc(
				"sbercdn_summary_traffic",
				"Float64 representing UNIX time in seconds since EPOCH since which certificate is valid",
				[]string{},
				nil), prometheus.GaugeValue},
		},
	}
}

func (c *SberCdnSummaryCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, m := range c.metrics {
		ch <- m.desc
	}
}

func (c *SberCdnSummaryCollector) Collect(mch chan<- prometheus.Metric) {
	t := time.Now().UTC()
	sst, err := c.client.GetSummaryStats(t)
	if err != nil {
		log.Panicln(err)
	}
	mch <- prometheus.MustNewConstMetric(
		c.metrics["sum_bw"].desc,
		c.metrics["sum_bw"].valueType,
		float64(sst.Bandwidth))
	mch <- prometheus.MustNewConstMetric(
		c.metrics["sum_ratio"].desc,
		c.metrics["sum_ratio"].valueType,
		sst.CacheRatio)
	mch <- prometheus.MustNewConstMetric(
		c.metrics["sum_hits"].desc,
		c.metrics["sum_hits"].valueType,
		float64(sst.Hits))
	mch <- prometheus.MustNewConstMetric(
		c.metrics["sum_traf"].desc,
		c.metrics["sum_traf"].valueType,
		float64(sst.Traffic))
}
