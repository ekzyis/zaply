package phoenixd

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/ekzyis/zaply/lightning"
)

type Phoenixd struct {
	url                *url.URL
	accessToken        string
	limitedAccessToken string
	webhookUrl         string
}

func NewPhoenixd(opts ...func(*Phoenixd) *Phoenixd) *Phoenixd {
	ln := &Phoenixd{}
	for _, opt := range opts {
		opt(ln)
	}
	return ln
}

func WithPhoenixdURL(u string) func(*Phoenixd) *Phoenixd {
	return func(p *Phoenixd) *Phoenixd {
		u, err := url.Parse(u)
		if err != nil {
			log.Fatal(err)
		}
		p.url = u
		return p
	}
}

func WithPhoenixdLimitedAccessToken(limitedAccessToken string) func(*Phoenixd) *Phoenixd {
	return func(p *Phoenixd) *Phoenixd {
		p.limitedAccessToken = limitedAccessToken
		return p
	}
}

func WithPhoenixdWebhookUrl(webhookUrl string) func(*Phoenixd) *Phoenixd {
	return func(p *Phoenixd) *Phoenixd {
		p.webhookUrl = webhookUrl
		return p
	}
}

func (p *Phoenixd) CreateInvoice(msats int64, description string) (lightning.Bolt11, error) {
	values := url.Values{}
	values.Add("amountSat", strconv.FormatInt(msats/1000, 10))
	values.Add("description", description)
	if p.webhookUrl != "" {
		values.Add("webhookUrl", p.webhookUrl)
	}

	endpoint := p.url.JoinPath("createinvoice")

	req, err := http.NewRequest("POST", endpoint.String(), strings.NewReader(values.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if p.limitedAccessToken != "" {
		req.SetBasicAuth("", p.limitedAccessToken)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("phoenixd %s: %s", resp.Status, string(body))
	}

	var response struct {
		Serialized string `json:"serialized"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return "", err
	}

	return lightning.Bolt11(response.Serialized), nil
}
