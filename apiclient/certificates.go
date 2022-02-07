package apiclient

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
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

func (c *SberCdnApiClient) GetCertList(account string) (certlist CertList, err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("failed to get certificates list:", r)
		}
	}()

	body, err := c.Get(
		fmt.Sprintf(c.endpoints["certList"], account),
		url.Values{})
	if err != nil {
		log.Panicln("failed to get certificates list:", err)
	}
	err = json.Unmarshal(body, &certlist)
	if err != nil {
		log.Panicf("failed to unmarshal cert list")
	}
	return certlist, err
}
