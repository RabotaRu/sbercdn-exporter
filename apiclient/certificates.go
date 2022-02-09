package apiclient

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"

	cmn "git.rabota.space/infrastructure/sbercdn-exporter/common"
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

func (c *SberCdnApiClient) GetCertList(ctx context.Context) (certlist CertList, err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("failed to get certificates list:", r)
		}
	}()
	var ok bool
	var account string
	if account, ok = ctx.Value(cmn.CtxKey("account")).(string); !ok {
		log.Panicln("no account value in context")
	}
	body, err := c.Get(
		fmt.Sprintf(c.endpoints["certList"], account),
		url.Values{},
		ctx)
	if err != nil {
		log.Panicln("failed to get certificates list:", err)
	}
	err = json.Unmarshal(body, &certlist)
	if err != nil {
		log.Panicf("failed to unmarshal cert list")
	}
	return certlist, err
}
