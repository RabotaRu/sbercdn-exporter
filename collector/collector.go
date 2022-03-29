package collector

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/RabotaRu/sbercdn-exporter/apiclient"
	cmn "github.com/RabotaRu/sbercdn-exporter/common"
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
	api_group     string
	client        *apiclient.SberCdnApiClient
	metrics       map[string]*Metric
	metric_names  map[string]string
	metric_groups []string
}

func NewSberCdnStatsCollector(client *apiclient.SberCdnApiClient) *SberCdnStatsCollector {
	col := SberCdnStatsCollector{
		api_group:     "statistic",
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
				metric_help+sep+metric_group_name, labels,
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
	var wag sync.WaitGroup
	ctx_root, ctx_cancel := context.WithTimeout(context.Background(), c.client.MaxQueryTime)
	defer ctx_cancel()
	ctx_root = context.WithValue(
		ctx_root,
		cmn.CtxKey("end"),
		time.Now().UTC().Truncate(time.Minute).Add(c.client.ScrapeTimeOffset*-1))

	sendMetrics := func(ctx context.Context) {
		defer wag.Done()
		var ok bool
		var acc, endpoint, group, group_name string
		var endtime time.Time
		var results []interface{}
		if acc, ok = ctx.Value(cmn.CtxKey("account")).(string); !ok {
			log.Println("Oops, acc is not a string!")
		}
		if endtime, ok = ctx.Value(cmn.CtxKey("end")).(time.Time); !ok {
			log.Println("Oops endtime is not of type time.Time!")
		}
		if group_name, ok = ctx.Value(cmn.CtxKey("metric_group_name")).(string); !ok {
			log.Println("Oops group_name is not a string!")
		}
		if group_name != SUMMARY {
			group = group_name
			endpoint = group + "s"
		}
		ctx = context.WithValue(context.WithValue(ctx, cmn.CtxKey("endpoint"), endpoint), cmn.CtxKey("api_group"), c.api_group)
		mtrc := c.client.GetStatistic(ctx) // c.api_group, endpoint, acc, endtime)
		if results, ok = mtrc["result"].([]interface{}); !ok {
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
				return
			}
		}

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

	for _, acc := range c.client.FindActiveAccounts(ctx_root) {
		ctx := context.WithValue(ctx_root, cmn.CtxKey("account"), acc)
		for _, metric_group_name := range c.metric_groups {
			wag.Add(1)
			ctx := context.WithValue(ctx, cmn.CtxKey("metric_group_name"), metric_group_name)
			go sendMetrics(ctx)
		}
	}
	wag.Wait()
}
