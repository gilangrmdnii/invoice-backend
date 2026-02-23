package request

type CreateInvoicePaymentRequest struct {
	InvoiceID     uint64  `json:"invoice_id" validate:"required"`
	Amount        float64 `json:"amount" validate:"required,gt=0"`
	PaymentDate   string  `json:"payment_date" validate:"required"`
	PaymentMethod string  `json:"payment_method" validate:"required,oneof=TRANSFER CASH GIRO OTHER"`
	ProofURL      string  `json:"proof_url" validate:"omitempty,max=500"`
	Notes         string  `json:"notes" validate:"max=2000"`
}
