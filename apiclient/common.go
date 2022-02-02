package apiclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	cmn "git.rabota.space/infrastructure/sbercdn-exporter/common"
)

const (
	TimeRangeFormat = "2006-01-02T15:04:05"
)

type SberCdnApiClient struct {
	*cmn.ClientConf
	hc              *http.Client
	auth_token_time time.Time
	endpoints       map[string]string
	auth_token      string
}

func NewSberCdnApiClient(options *cmn.ClientConf) (client *SberCdnApiClient, err error) {
	client = &SberCdnApiClient{
		hc:         &http.Client{},
		ClientConf: options,
	}
	_, err = client.auth()
	if err != nil {
		return nil, fmt.Errorf("initial authorization failed %w", err)
	}
	err = client.getAccountId()
	if err != nil {
		return nil, fmt.Errorf("failed to get account id: %w", err)
	}
	client.endpoints = map[string]string{
		"certList":  fmt.Sprintf("/app/ssl/v1/account/%v/certificate/", client.Auth.Id),
		"statistic": "/app/statistic/v3/",
	}
	return client, err
}

func (c *SberCdnApiClient) getAccountId() (err error) {
	body, err := c.Get("/app/inventory/v1/accounts/", url.Values{})
	if err != nil {
		return fmt.Errorf("failed to GET account_id")
	}
	var accs []map[string]string
	err = json.Unmarshal(body, &accs)
	if err != nil {
		return fmt.Errorf("failed to unmarshal accounts %w, \n\t%v", err, string(body))
	}
	if len(accs) == 0 {
		return fmt.Errorf("failed to find account_id, empty accounts list")
	}
	c.Auth.Id = strings.Split(accs[0]["uid"], "-")[2]
	log.Printf("using account: \"%v\"", c.Auth.Id)
	return
}

func (c *SberCdnApiClient) auth() (auth_token string, err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("failed to update auth token:", r)
			auth_token = ""
		}
	}()
	if c.auth_token != "" && time.Since(c.auth_token_time) < (c.Auth.TokenLifetime-c.MaxQueryTime) {
		return c.auth_token, err
	}
	req, err := http.NewRequestWithContext(
		context.Background(),
		"POST",
		c.Url+c.Auth.Urn,
		strings.NewReader(
			url.Values{
				"username": {c.Auth.Username},
				"password": {c.Auth.Password},
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

func (c *SberCdnApiClient) Get(urn string, query url.Values) (body []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			body = nil
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), c.ScrapeInterval)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", c.Url+urn, http.NoBody)
	if err != nil {
		log.Panicf("failed to prepare request for cert list: %v\n", err)
	}
	req.URL.RawQuery = query.Encode()
	auth_token, err := c.auth()
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

func (c *SberCdnApiClient) GetStatistic(endtime time.Time, endpoint string) (ms map[string]interface{}) {
	defer func() {
		if r := recover(); r != nil {
			ms = nil
		}
	}()
	v := url.Values{}
	v.Set("account", c.Auth.Id)
	v.Set("end", endtime.Truncate(time.Minute).Format(TimeRangeFormat))
	v.Set("start", endtime.Add(c.ScrapeInterval*-1).Truncate(time.Minute).Format(TimeRangeFormat))

	body, err := c.Get(c.endpoints["statistic"]+endpoint, v)
	if err != nil {
		log.Panicln("failed to query summary stats")
	}
	err = json.Unmarshal(body, &ms)
	if err != nil {
		log.Panicf("failed to unmarshal %v stats: %v", endpoint, err)
	}
	return ms
}