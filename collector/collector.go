package collector

import (
	"fmt"
	"sync"
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
			"bw": {prometheus.NewDesc(
				"sbercdn_bandwidth",
				"Peak bandwidth total",
				[]string{},
				nil), prometheus.GaugeValue},
			"ratio": {prometheus.NewDesc(
				"sbercdn_cache_ration",
				"Cache hit ration total",
				[]string{},
				nil), prometheus.GaugeValue},
			"hits": {prometheus.NewDesc(
				"sbercdn_hits",
				"Cache hits total",
				[]string{},
				nil), prometheus.GaugeValue},
			"traf": {prometheus.NewDesc(
				"sbercdn_traffic",
				"Traffic total",
				[]string{},
				nil), prometheus.GaugeValue},
			"code_bw": {prometheus.NewDesc(
				"sbercdn_code_bandwidth",
				"Peak bandwidth by http code",
				[]string{"code"},
				nil), prometheus.GaugeValue},
			"code_ratio": {prometheus.NewDesc(
				"sbercdn_code_cache_ration",
				"Cache hit ration by http code",
				[]string{"code"},
				nil), prometheus.GaugeValue},
			"code_hits": {prometheus.NewDesc(
				"sbercdn_code_hits",
				"Cache hits by http code",
				[]string{"code"},
				nil), prometheus.GaugeValue},
			"code_traf": {prometheus.NewDesc(
				"sbercdn_code_traffic",
				"Traffic by http code",
				[]string{"code"},
				nil), prometheus.GaugeValue},
			"res_bw": {prometheus.NewDesc(
				"sbercdn_resource_bandwidth",
				"Peak bandwidth by resource name",
				[]string{"resource_name"},
				nil), prometheus.GaugeValue},
			"res_ratio": {prometheus.NewDesc(
				"sbercdn_resource_cache_ration",
				"Cache hit ration by resource name",
				[]string{"resource_name"},
				nil), prometheus.GaugeValue},
			"res_hits": {prometheus.NewDesc(
				"sbercdn_resource_hits",
				"Cache hits by resource name",
				[]string{"resource_name"},
				nil), prometheus.GaugeValue},
			"res_traf": {prometheus.NewDesc(
				"sbercdn_resource_traffic",
				"Traffic by resource name",
				[]string{"resource_name"},
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
	endtime := time.Now().UTC()
	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()
		mtrc := c.client.GetMetrics(endtime, "summary")
		if len(mtrc) == 0 {
			return
		}
		mch <- prometheus.MustNewConstMetric(
			c.metrics["bw"].desc,
			c.metrics["bw"].valueType,
			mtrc["bandwidth"].(float64))
		mch <- prometheus.MustNewConstMetric(
			c.metrics["ratio"].desc,
			c.metrics["ratio"].valueType,
			mtrc["cache_ratio"].(float64))
		mch <- prometheus.MustNewConstMetric(
			c.metrics["hits"].desc,
			c.metrics["hits"].valueType,
			mtrc["hits"].(float64))
		mch <- prometheus.MustNewConstMetric(
			c.metrics["traf"].desc,
			c.metrics["traf"].valueType,
			mtrc["traffic"].(float64))
	}()

	go func() {
		defer wg.Done()
		mtrc := c.client.GetMetrics(endtime, "codes")
		if len(mtrc) == 0 {
			return
		}
		for _, v := range mtrc["result"].([]interface{}) {
			met, ok := v.(map[string]interface{})
			if !ok {
				return
			}
			mch <- prometheus.MustNewConstMetric(
				c.metrics["code_bw"].desc,
				c.metrics["code_bw"].valueType,
				met["bandwidth"].(float64), fmt.Sprintf("%v", met["code"].(float64)))
			mch <- prometheus.MustNewConstMetric(
				c.metrics["code_ratio"].desc,
				c.metrics["code_ratio"].valueType,
				met["cache_ratio"].(float64), fmt.Sprintf("%v", met["code"].(float64)))
			mch <- prometheus.MustNewConstMetric(
				c.metrics["code_hits"].desc,
				c.metrics["code_hits"].valueType,
				met["hits"].(float64), fmt.Sprintf("%v", met["code"].(float64)))
			mch <- prometheus.MustNewConstMetric(
				c.metrics["code_traf"].desc,
				c.metrics["code_traf"].valueType,
				met["traffic"].(float64), fmt.Sprintf("%v", met["code"].(float64)))
		}
	}()

	go func() {
		defer wg.Done()
		mtrc := c.client.GetMetrics(endtime, "resources")
		if len(mtrc) == 0 {
			return
		}
		for _, v := range mtrc["result"].([]interface{}) {
			met, ok := v.(map[string]interface{})
			if !ok {
				return
			}
			mch <- prometheus.MustNewConstMetric(
				c.metrics["res_bw"].desc,
				c.metrics["res_bw"].valueType,
				met["bandwidth"].(float64), met["resource_name"].(string))
			mch <- prometheus.MustNewConstMetric(
				c.metrics["res_ratio"].desc,
				c.metrics["res_ratio"].valueType,
				met["cache_ratio"].(float64), met["resource_name"].(string))
			mch <- prometheus.MustNewConstMetric(
				c.metrics["res_hits"].desc,
				c.metrics["res_hits"].valueType,
				met["hits"].(float64), met["resource_name"].(string))
			mch <- prometheus.MustNewConstMetric(
				c.metrics["res_traf"].desc,
				c.metrics["res_traf"].valueType,
				met["traffic"].(float64), met["resource_name"].(string))
		}
	}()

	wg.Wait()
}
