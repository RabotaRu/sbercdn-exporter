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
			"bandwidth": {prometheus.NewDesc(
				"sbercdn_bandwidth",
				"Peak bandwidth total",
				[]string{},
				nil), prometheus.GaugeValue},
			"cache_ratio": {prometheus.NewDesc(
				"sbercdn_cache_ration",
				"Cache hit ration total",
				[]string{},
				nil), prometheus.GaugeValue},
			"hits": {prometheus.NewDesc(
				"sbercdn_hits",
				"Cache hits total",
				[]string{},
				nil), prometheus.GaugeValue},
			"traffic": {prometheus.NewDesc(
				"sbercdn_traffic",
				"Traffic total",
				[]string{},
				nil), prometheus.GaugeValue},
			"code_bandwidth": {prometheus.NewDesc(
				"sbercdn_code_bandwidth",
				"Peak bandwidth by http code",
				[]string{"code"},
				nil), prometheus.GaugeValue},
			"code_cache_ratio": {prometheus.NewDesc(
				"sbercdn_code_cache_ration",
				"Cache hit ration by http code",
				[]string{"code"},
				nil), prometheus.GaugeValue},
			"code_hits": {prometheus.NewDesc(
				"sbercdn_code_hits",
				"Cache hits by http code",
				[]string{"code"},
				nil), prometheus.GaugeValue},
			"code_traffic": {prometheus.NewDesc(
				"sbercdn_code_traffic",
				"Traffic by http code",
				[]string{"code"},
				nil), prometheus.GaugeValue},
			"resource_bandwidth": {prometheus.NewDesc(
				"sbercdn_resource_bandwidth",
				"Peak bandwidth by resource name",
				[]string{"resource_name"},
				nil), prometheus.GaugeValue},
			"resource_cache_ratio": {prometheus.NewDesc(
				"sbercdn_resource_cache_ration",
				"Cache hit ration by resource name",
				[]string{"resource_name"},
				nil), prometheus.GaugeValue},
			"resource_hits": {prometheus.NewDesc(
				"sbercdn_resource_hits",
				"Cache hits by resource name",
				[]string{"resource_name"},
				nil), prometheus.GaugeValue},
			"resource_traffic": {prometheus.NewDesc(
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
	endtime := time.Now().UTC().Truncate(time.Minute).Add(time.Minute * -5)
	var wag sync.WaitGroup
	wag.Add(3)

	go func() {
		defer wag.Done()
		mtrc := c.client.GetMetrics(endtime, "summary")
		if len(mtrc) == 0 {
			return
		}
		for key, v := range mtrc {
			if mtv, ok := v.(float64); ok {
				mch <- prometheus.NewMetricWithTimestamp(
					endtime,
					prometheus.MustNewConstMetric(
						c.metrics[key].desc,
						c.metrics[key].valueType,
						mtv))
			}
		}
	}()

	sendMetrics := func(grp string) {
		defer wag.Done()
		mtrc := c.client.GetMetrics(endtime, grp+"s")
		if len(mtrc) == 0 {
			return
		}
		var ok bool
		var result []interface{}
		if result, ok = mtrc["result"].([]interface{}); !ok {
			return
		}
		for _, v := range result {
			met, ok := v.(map[string]interface{})
			if !ok {
				return
			}
			var label interface{}
			if label, ok = met[grp+"_name"]; ok {
				delete(met, grp+"_name")
			} else {
				label = met[grp]
			}
			delete(met, grp)
			for key, v := range met {
				mtnm := grp + "_" + key
				if mtv, ok := v.(float64); ok {
					mch <- prometheus.NewMetricWithTimestamp(
						endtime,
						prometheus.MustNewConstMetric(
							c.metrics[mtnm].desc,
							c.metrics[mtnm].valueType,
							mtv, fmt.Sprintf("%v", label)))
				}
			}
		}
	}
	go sendMetrics("code")
	go sendMetrics("resource")

	wag.Wait()
}
