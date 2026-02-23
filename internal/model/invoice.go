package model

import "time"

type InvoiceType string

const (
	InvoiceTypeDP           InvoiceType = "DP"
	InvoiceTypeFinalPayment InvoiceType = "FINAL_PAYMENT"
	InvoiceTypeTOP1         InvoiceType = "TOP_1"
	InvoiceTypeTOP2         InvoiceType = "TOP_2"
	InvoiceTypeTOP3         InvoiceType = "TOP_3"
	InvoiceTypeMeals        InvoiceType = "MEALS"
	InvoiceTypeAdditional   InvoiceType = "ADDITIONAL"
)

type InvoiceStatus string

const (
	InvoiceStatusPending  InvoiceStatus = "PENDING"
	InvoiceStatusApproved InvoiceStatus = "APPROVED"
	InvoiceStatusRejected InvoiceStatus = "REJECTED"
)

type Invoice struct {
	ID               uint64        `json:"id"`
	InvoiceNumber    string        `json:"invoice_number"`
	InvoiceType      InvoiceType   `json:"invoice_type"`
	ProjectID        uint64        `json:"project_id"`
	Amount           float64       `json:"amount"`
	PaidAmount       float64       `json:"paid_amount"`
	Status           InvoiceStatus `json:"status"`
	PaymentStatus    PaymentStatus `json:"payment_status"`
	FileURL          string        `json:"file_url,omitempty"`
	RecipientName    string        `json:"recipient_name"`
	RecipientAddress string        `json:"recipient_address,omitempty"`
	Attention        string        `json:"attention,omitempty"`
	PONumber         string        `json:"po_number,omitempty"`
	InvoiceDate      time.Time     `json:"invoice_date"`
	DueDate          *time.Time    `json:"due_date,omitempty"`
	DPPercentage     *float64      `json:"dp_percentage,omitempty"`
	Subtotal         float64       `json:"subtotal"`
	TaxPercentage    float64       `json:"tax_percentage"`
	TaxAmount        float64       `json:"tax_amount"`
	Notes            string        `json:"notes,omitempty"`
	Language         string        `json:"language"`
	CreatedBy        uint64        `json:"created_by"`
	ApprovedBy       *uint64       `json:"approved_by,omitempty"`
	RejectNotes      string        `json:"reject_notes,omitempty"`
	CreatedAt        time.Time     `json:"created_at"`
	UpdatedAt        time.Time     `json:"updated_at"`
}
