package apiclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	cmn "git.rabota.space/infrastructure/sbercdn-exporter/common"
)

const (
	TimeRangeFormat = "2006-01-02T15:04:05"
)

func InArray(a []string, e string) bool {
	for _, x := range a {
		if x == e {
			return true
		}
	}
	return false
}

type SberCdnApiClient struct {
	*cmn.ClientConf
	hc               *http.Client
	auth_token_time  time.Time
	accs_update_time *time.Time
	endpoints        map[string]string
	auth_token       string
	active_accs      []string
}

func NewSberCdnApiClient(options *cmn.ClientConf) (client *SberCdnApiClient, err error) {
	client = &SberCdnApiClient{
		hc:         &http.Client{},
		ClientConf: options,
	}
	sort.Strings(client.Accounts)
	ctx := context.Background()
	_, err = client.auth(ctx)
	if err != nil {
		return nil, fmt.Errorf("initial authorization failed %w", err)
	}
	client.FindActiveAccounts(ctx)
	client.endpoints = map[string]string{
		"certList":     "/app/ssl/v1/account/%v/certificate/",
		"statistic":    "/app/statistic/v3/",
		"realtimestat": "/app/realtimestat/v1/",
	}
	return client, err
}

func (c *SberCdnApiClient) FindActiveAccounts(ctx context.Context) (accounts []string) {
	defer func() {
		if r := recover(); r != nil {
			log.Println(r)
		}
	}()
	body, err := c.Get("/app/inventory/v1/accounts/", url.Values{}, ctx)
	if err != nil {
		log.Panicln("failed to GET account_id")
	}
	var accs []map[string]string
	err = json.Unmarshal(body, &accs)
	if err != nil {
		log.Panicln("failed to unmarshal accounts %w", err)
	}
	if len(accs) == 0 {
		log.Panicln("failed to find account_id, empty accounts list")
	}
	if c.accs_update_time == nil || time.Since(*c.accs_update_time) >= c.ScrapeTimeOffset {
		var active_accs []string
		for i := range accs {
			if accs[i]["status"] == "active" && (len(c.Accounts) == 0 || InArray(c.Accounts, accs[i]["name"])) {
				active_accs = append(active_accs, accs[i]["name"])
			}
		}
		c.active_accs = active_accs
		t := time.Now()
		c.accs_update_time = &t
	}
	return c.active_accs
}

func (c *SberCdnApiClient) auth(ctx context.Context) (auth_token string, err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("failed to update auth token:", r)
			auth_token = ""
		}
	}()
	if c.auth_token != "" && time.Since(c.auth_token_time) < (c.TokenLifetime-c.MaxQueryTime) {
		return c.auth_token, err
	}
	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		c.Url+c.AuthUrn,
		strings.NewReader(
			url.Values{
				"username": {c.Username},
				"password": {c.Password},
			}.Encode(),
		),
	)
	if err != nil {
		log.Panicln("Error creating new auth request")
	}
	resp, err := c.hc.Do(req)
	if err != nil {
		log.Panicln("Error sending auth request")
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Panicln(fmt.Errorf("auth response status code %v", resp.Status))
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Panicln("Error reading auth response body")
	}

	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Panicln("Error unmarshaling auth response json body")
	}

	if auth_token, ok := data["token"].(string); ok {
		c.auth_token = auth_token
		c.auth_token_time = time.Now()
		log.Println("Authorized successfully!")
	} else {
		err := fmt.Errorf("token is not a string")
		log.Panicln("%w", err)
	}
	return c.auth_token, err
}

func (c *SberCdnApiClient) Get(urn string, query url.Values, ctx context.Context) (body []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			body = nil
		}
	}()

	// ctx, cancel := context.WithTimeout(ctx, c.ScrapeInterval)
	// defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", c.Url+urn, http.NoBody)
	if err != nil {
		log.Panicf("failed to prepare request for cert list: %v\n", err)
	}
	req.URL.RawQuery = query.Encode()
	auth_token, err := c.auth(ctx)
	if err != nil {
		log.Panicln(err)
	}
	req.Header.Add("Cdn-Auth-Token", auth_token)
	resp, err := c.hc.Do(req)
	if err != nil {
		log.Panicf("failed to send request for cert list %v\n", err)
	}
	defer resp.Body.Close()
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		log.Panicf("faile to read response body for cert list %v\n", err)
	}
	return body, err
}

// func (c *SberCdnApiClient) GetStatistic(api_group, endpoint, account string, endtime time.Time) (ms map[string]interface{}) {
func (c *SberCdnApiClient) GetStatistic(ctx context.Context) (ms map[string]interface{}) {
	defer func() {
		if r := recover(); r != nil {
			ms = nil
		}
	}()
	var uri string
	values := url.Values{}
	if account, ok := ctx.Value(cmn.CtxKey("account")).(string); ok {
		values.Set("account", account)
	} else {
		log.Panicln("no account in context")
	}
	if end, ok := ctx.Value(cmn.CtxKey("end")).(time.Time); ok {
		values.Set("end", end.Format(TimeRangeFormat))
		values.Set("start", end.Add(c.ScrapeInterval*-1).Format(TimeRangeFormat))
	} else {
		log.Panicln("no request time in context")
	}
	if api_group, ok := ctx.Value(cmn.CtxKey("api_group")).(string); ok {
		uri = c.endpoints[api_group]
	} else {
		log.Println("no api_group in context")
	}
	if endpoint, ok := ctx.Value(cmn.CtxKey("endpoint")).(string); ok {
		uri += endpoint
	} else {
		log.Panicln("no api endpoint in context")
	}
	body, err := c.Get(uri, values, ctx)
	if err != nil {
		log.Panicln("failed to query summary stats")
	}
	err = json.Unmarshal(body, &ms)
	if err != nil {
		log.Panicf("failed to unmarshal %v: %v", uri, err)
	}
	return ms
}
