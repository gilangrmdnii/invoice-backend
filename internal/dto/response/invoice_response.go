package response

import "time"

type InvoiceItemResponse struct {
	ID          uint64  `json:"id"`
	InvoiceID   uint64  `json:"invoice_id"`
	Description string  `json:"description"`
	Quantity    float64 `json:"quantity"`
	Unit        string  `json:"unit"`
	UnitPrice   float64 `json:"unit_price"`
	Subtotal    float64 `json:"subtotal"`
	SortOrder   int     `json:"sort_order"`
}

type InvoiceResponse struct {
	ID               uint64                `json:"id"`
	InvoiceNumber    string                `json:"invoice_number"`
	InvoiceType      string                `json:"invoice_type"`
	ProjectID        uint64                `json:"project_id"`
	Amount           float64               `json:"amount"`
	Status           string                `json:"status"`
	FileURL          string                `json:"file_url,omitempty"`
	RecipientName    string                `json:"recipient_name"`
	RecipientAddress string                `json:"recipient_address,omitempty"`
	Attention        string                `json:"attention,omitempty"`
	PONumber         string                `json:"po_number,omitempty"`
	InvoiceDate      string                `json:"invoice_date"`
	DPPercentage     *float64              `json:"dp_percentage,omitempty"`
	Subtotal         float64               `json:"subtotal"`
	TaxPercentage    float64               `json:"tax_percentage"`
	TaxAmount        float64               `json:"tax_amount"`
	Notes            string                `json:"notes,omitempty"`
	Language         string                `json:"language"`
	CreatedBy        uint64                `json:"created_by"`
	ApprovedBy       *uint64               `json:"approved_by,omitempty"`
	RejectNotes      string                `json:"reject_notes,omitempty"`
	Items            []InvoiceItemResponse `json:"items"`
	CreatedAt        time.Time             `json:"created_at"`
	UpdatedAt        time.Time             `json:"updated_at"`
}
