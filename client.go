package main

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
)

type CertItem struct {
	Alt     string  `json:"alt"`
	Cn      string  `json:"cn"`
	Issuer  string  `json:"issuer"`
	Comment string  `json:"comment"`
	Start   float64 `json:"start"`
	End     float64 `json:"end"`
}

type CertList struct {
	Message string     `json:"message"`
	Data    []CertItem `json:"data"`
	Status  int        `json:"status"`
}

type SberCdnApiClient struct {
	hc              *http.Client
	conf            *ClientConf
	auth_token_time time.Time
	auth_token      string
}

func NewSberCdnApiClient(conf *ClientConf) *SberCdnApiClient {
	return &SberCdnApiClient{
		hc:   &http.Client{},
		conf: conf,
	}
}

func (ac *SberCdnApiClient) auth() (auth_token string, err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("failed to update auth token:", r)
		}
	}()
	if ac.auth_token != "" && time.Since(ac.auth_token_time) < (ac.conf.Auth.TokenLifetime-ac.conf.MaxQueryTime) {
		return ac.auth_token, err
	}
	req, err := http.NewRequestWithContext(
		context.Background(),
		"POST",
		ac.conf.ApiUrl+ac.conf.Auth.Urn,
		strings.NewReader(
			url.Values{
				"username": {ac.conf.Auth.Username},
				"password": {ac.conf.Auth.Password},
			}.Encode(),
		),
	)
	if err != nil {
		log.Panicln("Error creating new auth request")
	}
	resp, err := ac.hc.Do(req)
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

	var um map[string]interface{}
	err = json.Unmarshal(body, &um)
	if err != nil {
		log.Panicln("Error unmarshaling auth response json body")
	}

	if auth_token, ok := um["token"].(string); ok {
		ac.auth_token = auth_token
		ac.auth_token_time = time.Now()
	} else {
		err := fmt.Errorf("token is not a string")
		log.Panicln("%w", err)
	}
	return ac.auth_token, err
}

func (ac *SberCdnApiClient) GetCertList() (certlist []CertItem, err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("failed to get certificates list:", r)
		}
	}()

	req, err := http.NewRequestWithContext(
		context.Background(),
		"GET",
		ac.conf.ApiUrl+strings.ReplaceAll(ac.conf.URNs.CertList, "{{ auth.id }}", ac.conf.Auth.Id),
		http.NoBody)
	if err != nil {
		log.Panicf("failed to prepare request for cert list: %v\n", err)
	}
	auth_token, err := ac.auth()
	if err != nil {
		log.Panicln(err)
	}
	req.Header.Add("Cdn-Auth-Token", auth_token)
	resp, err := ac.hc.Do(req)
	if err != nil {
		log.Panicf("failed to send request for cert list %v\n", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Panicf("faile to read response body for cert list %v\n", err)
	}

	var result CertList
	err = json.Unmarshal(body, &result)
	if err != nil {
		log.Panicf("failed to unmarshal cert list")
	}
	return result.Data, err
}
