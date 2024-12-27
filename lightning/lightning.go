package lightning

import "time"

type PaymentRequest string

type Lightning interface {
	CreateInvoice(msats int64, description string) (PaymentRequest, error)
	GetInvoice(paymentHash string) (*Invoice, error)

	IncomingPayments() chan *Invoice
}

type Invoice struct {
	PaymentHash    string    `json:"paymentHash"`
	Preimage       string    `json:"preimage"`
	Msats          int64     `json:"msats"`
	Description    string    `json:"description"`
	PaymentRequest string    `json:"paymentRequest"`
	CreatedAt      time.Time `json:"createdAt"`
	ConfirmedAt    time.Time `json:"confirmedAt"`
}
