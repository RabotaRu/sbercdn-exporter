package api_client

import (
	"encoding/json"
	"log"
	"net/url"
	"time"
)

const (
	TimeRangeFormat = "2006-01-02T15:04:05"
)

type Stats struct {
	Hits       float64 `json:"hits"`        // 2104772,
	Traffic    float64 `json:"traffic"`     // 40603470765,
	CacheRatio float64 `json:"cache_ratio"` // 0.8231922443391645,
	Bandwidth  float64 `json:"bandwidth"`   // 5612541
}

type ResourceStats struct {
	Resource     string `json:"resource"`
	ResourceName string `json:"resource_name"`
	Stats
}

type CountryStats struct {
	Country     string `json:"country"`
	CountryName string `json:"country_name"`
	Stats
}

type StatsRequestParams struct {
	Account string `json:"account"`
	Start   string `json:"start"` // "2019-08-01T11:01:00",
	End     string `json:"end"`   // "2019-08-19T10:01:00",
	Code    string `json:"code"`  // "200"
}

type CodeResult struct {
	Code int16 `json:"code"`
	Stats
}

type SummaryStats struct {
	StatsRequestParams
	Stats
}

func (c *SberCdnApiClient) GetSummaryStats(end time.Time) (stats *SummaryStats, err error) {
	defer func() {
		if r := recover(); r != nil {
			stats = nil
		}
	}()
	v := url.Values{}
	v.Set("end", end.Truncate(time.Minute).Format(TimeRangeFormat))
	v.Set("start", end.Add(c.ScrapeInterval*-1).Truncate(time.Minute).Format(TimeRangeFormat))

	body, err := c.Get("/app/statistic/v3/", v)
	if err != nil {
		log.Panicln("failed to query summary stats")
	}
	err = json.Unmarshal(body, &stats)
	if err != nil {
		log.Panicf("failed to unmarshal summary stats")
	}
	log.Printf("%+v\n", stats)
	return stats, err
}
