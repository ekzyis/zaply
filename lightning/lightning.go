package lightning

import "time"

type PaymentRequest string

type Lightning interface {
	CreateInvoice(msats int64, description string) (PaymentRequest, error)
	GetInvoice(paymentHash string) (*Invoice, error)
}

type Invoice struct {
	PaymentHash    string
	Preimage       string
	Msats          int64
	Description    string
	PaymentRequest string
	CreatedAt      time.Time
	ConfirmedAt    time.Time
}
