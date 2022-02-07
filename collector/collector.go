package collector

import (
	"fmt"
	"sync"
	"time"

	"git.rabota.space/infrastructure/sbercdn-exporter/apiclient"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	NAMESPACE = "sbercdn"
	SUMMARY   = "summary"
)

type Metric struct {
	desc      *prometheus.Desc
	valueType prometheus.ValueType
}

type SberCdnStatsCollector struct {
	client        *apiclient.SberCdnApiClient
	metrics       map[string]*Metric
	metric_names  map[string]string
	metric_groups []string
}

func NewSberCdnStatsCollector(client *apiclient.SberCdnApiClient) *SberCdnStatsCollector {
	col := SberCdnStatsCollector{
		client:        client,
		metrics:       make(map[string]*Metric),
		metric_groups: []string{SUMMARY, "code", "resource"},
		metric_names: map[string]string{
			"bandwidth":   "Peak bandwidth in bits",
			"cache_ratio": "Cache hit ratio",
			"hits":        "Cache hits",
			"traffic":     "Traffic in bytes",
		},
	}

	for _, metric_group_name := range col.metric_groups {
		var stats_group string
		sep := " "
		labels := []string{"account"}
		if metric_group_name != SUMMARY {
			labels = append(labels, metric_group_name)
			stats_group = metric_group_name
			sep = " by "
		}
		for metric_name, metric_help := range col.metric_names {
			col.metrics[prometheus.BuildFQName("", stats_group, metric_name)] = &Metric{prometheus.NewDesc(
				prometheus.BuildFQName(NAMESPACE, stats_group, metric_name),
				metric_help + sep + metric_group_name, labels,
				nil), prometheus.GaugeValue}
		}
	}
	return &col
}

func (c *SberCdnStatsCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, m := range c.metrics {
		ch <- m.desc
	}
}

func (c *SberCdnStatsCollector) Collect(mch chan<- prometheus.Metric) {
	endtime := time.Now().UTC().Truncate(time.Minute).Add(c.client.ScrapeTimeOffset * -1)
	var wag sync.WaitGroup

	sendMetrics := func(acc, group_name string) {
		defer wag.Done()
		var endpoint, group string
		if group_name != SUMMARY {
			group = group_name
			endpoint = group + "s"
		}
		mtrc := c.client.GetStatistic(endtime, endpoint, acc)
		if results, ok := mtrc["result"].([]interface{}); !ok {
			if group_name == SUMMARY {
				for metric_name, value := range mtrc {
					if metric_value, ok := value.(float64); ok {
						mch <- prometheus.NewMetricWithTimestamp(
							endtime,
							prometheus.MustNewConstMetric(
								c.metrics[metric_name].desc,
								c.metrics[metric_name].valueType,
								metric_value, acc))
					}
				}
			}
		} else {
			for _, result_value := range results {
				if metrics, ok := result_value.(map[string]interface{}); ok {
					var label interface{}
					if group_name != SUMMARY {
						if label, ok = metrics[group+"_name"]; ok {
							delete(metrics, group+"_name")
						} else {
							label = metrics[group]
						}
						delete(metrics, group)
					}
					for key, v := range metrics {
						metric_name := group + "_" + key
						if metric_value, ok := v.(float64); ok {
							mch <- prometheus.NewMetricWithTimestamp(
								endtime,
								prometheus.MustNewConstMetric(
									c.metrics[metric_name].desc,
									c.metrics[metric_name].valueType,
									metric_value, acc, fmt.Sprintf("%v", label)))
						}
					}
				}
			}
		}
	}

	for _, acc := range c.client.FindActiveAccounts() {
		for _, metric_group_name := range c.metric_groups {
			wag.Add(1)
			go sendMetrics(acc, metric_group_name)
		}
	}
	wag.Wait()
}
