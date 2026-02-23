package model

import "time"

type PaymentMethod string

const (
	PaymentMethodTransfer PaymentMethod = "TRANSFER"
	PaymentMethodCash     PaymentMethod = "CASH"
	PaymentMethodGiro     PaymentMethod = "GIRO"
	PaymentMethodOther    PaymentMethod = "OTHER"
)

type PaymentStatus string

const (
	PaymentStatusUnpaid      PaymentStatus = "UNPAID"
	PaymentStatusPartialPaid PaymentStatus = "PARTIAL_PAID"
	PaymentStatusPaid        PaymentStatus = "PAID"
)

type InvoicePayment struct {
	ID            uint64        `json:"id"`
	InvoiceID     uint64        `json:"invoice_id"`
	Amount        float64       `json:"amount"`
	PaymentDate   time.Time     `json:"payment_date"`
	PaymentMethod PaymentMethod `json:"payment_method"`
	ProofURL      string        `json:"proof_url,omitempty"`
	Notes         string        `json:"notes,omitempty"`
	CreatedBy     uint64        `json:"created_by"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}
