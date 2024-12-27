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
	"time"

	"github.com/ekzyis/zaply/lightning"
	"github.com/labstack/echo/v4"
)

type Phoenixd struct {
	url                *url.URL
	accessToken        string
	limitedAccessToken string
	webhookUrl         string

	paymentsChan chan *lightning.Invoice
}

func NewPhoenixd(opts ...func(*Phoenixd) *Phoenixd) *Phoenixd {
	ln := &Phoenixd{
		paymentsChan: make(chan *lightning.Invoice),
	}
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

func (p *Phoenixd) CreateInvoice(msats int64, description string) (lightning.PaymentRequest, error) {
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

	return lightning.PaymentRequest(response.Serialized), nil
}

func (p *Phoenixd) GetInvoice(paymentHash string) (*lightning.Invoice, error) {
	endpoint := p.url.JoinPath("payments/incoming", paymentHash)

	req, err := http.NewRequest("GET", endpoint.String(), nil)
	if err != nil {
		return nil, err
	}
	if p.limitedAccessToken != "" {
		req.SetBasicAuth("", p.limitedAccessToken)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("phoenixd %s: %s", resp.Status, string(body))
	}

	var response struct {
		PaymentHash string `json:"paymentHash"`
		Preimage    string `json:"preimage"`
		Sats        int64  `json:"receivedSat"`
		Description string `json:"description"`
		CreatedAt   int64  `json:"createdAt"`
		ConfirmedAt int64  `json:"completedAt"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	createdAt := time.Unix(response.CreatedAt/1000, 0)
	var confirmedAt time.Time
	if response.ConfirmedAt != 0 {
		confirmedAt = time.Unix(response.ConfirmedAt/1000, 0)
	}

	return &lightning.Invoice{
		PaymentHash: response.PaymentHash,
		Preimage:    response.Preimage,
		Msats:       response.Sats * 1_000,
		Description: response.Description,
		CreatedAt:   createdAt,
		ConfirmedAt: confirmedAt,
	}, nil
}

func (p *Phoenixd) IncomingPayments() chan *lightning.Invoice {
	return p.paymentsChan
}

func (p *Phoenixd) WebhookHandler(c echo.Context) error {
	var webhook struct {
		Type        string `json:"type"`
		AmountSat   int64  `json:"amountSat"`
		PaymentHash string `json:"paymentHash"`
	}
	if err := c.Bind(&webhook); err != nil {
		return err
	}

	go func() {
		inv, err := p.GetInvoice(webhook.PaymentHash)
		if err != nil {
			c.Logger().Error(err)
			return
		}

		log.Printf(
			"payment received: %s | %d msats | %s | %s | %s",
			inv.PaymentHash, inv.Msats, inv.Description,
			inv.CreatedAt.Format(time.RFC3339), inv.ConfirmedAt.Format(time.RFC3339),
		)

		p.paymentsChan <- inv
	}()

	return c.NoContent(http.StatusOK)
}
