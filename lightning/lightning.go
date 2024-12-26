package lightning

type Bolt11 string

type Lightning interface {
	CreateInvoice(msats int64, description string) (Bolt11, error)
}
