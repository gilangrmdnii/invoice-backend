package response

import "time"

type InvoicePaymentResponse struct {
	ID            uint64    `json:"id"`
	InvoiceID     uint64    `json:"invoice_id"`
	Amount        float64   `json:"amount"`
	PaymentDate   string    `json:"payment_date"`
	PaymentMethod string    `json:"payment_method"`
	ProofURL      string    `json:"proof_url,omitempty"`
	Notes         string    `json:"notes,omitempty"`
	CreatedBy     uint64    `json:"created_by"`
	CreatorName   string    `json:"creator_name,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}
