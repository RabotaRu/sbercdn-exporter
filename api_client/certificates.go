package api_client

import (
	"encoding/json"
	"log"
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

func (ac *SberCdnApiClient) GetCertList() (certlist []CertItem, err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("failed to get certificates list:", r)
		}
	}()

	body, err := ac.Get("CertList")
	if err != nil {
		log.Panicln("failed to get certificates list: %v", err)
	}
	var result CertList
	err = json.Unmarshal(body, &result)
	if err != nil {
		log.Panicf("failed to unmarshal cert list")
	}
	return result.Data, err
}
