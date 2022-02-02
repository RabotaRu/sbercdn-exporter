package collector

import (
	"fmt"
	"sync"
	"time"

	"git.rabota.space/infrastructure/sbercdn-exporter/api_client"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "sbercdn"
	summary   = "summary"
)

type Metric struct {
	desc      *prometheus.Desc
	valueType prometheus.ValueType
}

type SberCdnSummaryCollector struct {
	client        *api_client.SberCdnApiClient
	metrics       map[string]*Metric
	metric_names  map[string]string
	metric_groups []string
}

func NewSberCdnSummaryCollector(client *api_client.SberCdnApiClient) *SberCdnSummaryCollector {
	col := SberCdnSummaryCollector{
		client:        client,
		metrics:       make(map[string]*Metric),
		metric_groups: []string{summary, "code", "resource"},
		metric_names: map[string]string{
			"bandwidth":   "Peak bandwidth in bits",
			"cache_ratio": "Cache hit ratio",
			"hits":        "Cache hits",
			"traffic":     "Traffic in bytes",
		},
	}

	for _, metric_group_name := range col.metric_groups {
		for metric_name, metric_help := range col.metric_names {
			var label_name, help string
			var labels []string
			if metric_group_name == summary {
				help = metric_help + " " + metric_group_name
			} else {
				help = metric_help + " by " + metric_group_name
				label_name = metric_group_name
				labels = []string{label_name}
			}
			col.metrics[prometheus.BuildFQName("", label_name, metric_name)] = &Metric{prometheus.NewDesc(
				prometheus.BuildFQName(namespace, label_name, metric_name),
				help, labels,
				nil), prometheus.GaugeValue}
		}
	}
	return &col
}

func (c *SberCdnSummaryCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, m := range c.metrics {
		ch <- m.desc
	}
}

func (c *SberCdnSummaryCollector) Collect(mch chan<- prometheus.Metric) {
	endtime := time.Now().UTC().Truncate(time.Minute).Add(time.Minute * -10)
	var wag sync.WaitGroup

	sendMetrics := func(group_name string) {
		defer wag.Done()
		var endpoint, group string
		if group_name != summary {
			group = group_name
			endpoint = group + "s"
		}
		mtrc := c.client.GetStatistic(endtime, endpoint)
		if results, ok := mtrc["result"].([]interface{}); !ok {
			if group_name == summary {
				for metric_name, value := range mtrc {
					if metric_value, ok := value.(float64); ok {
						mch <- prometheus.NewMetricWithTimestamp(
							endtime,
							prometheus.MustNewConstMetric(
								c.metrics[metric_name].desc,
								c.metrics[metric_name].valueType,
								metric_value))
					}
				}
			}
		} else {
			for _, result_value := range results {
				if metrics, ok := result_value.(map[string]interface{}); ok {
					var label interface{}
					if group_name != summary {
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
									metric_value, fmt.Sprintf("%v", label)))
						}
					}
				}
			}
		}
	}

	for _, metric_group_name := range c.metric_groups {
		wag.Add(1)
		go sendMetrics(metric_group_name)
	}

	wag.Wait()
}
