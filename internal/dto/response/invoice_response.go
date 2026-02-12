package response

import "time"

type InvoiceResponse struct {
	ID            uint64    `json:"id"`
	InvoiceNumber string    `json:"invoice_number"`
	ProjectID     uint64    `json:"project_id"`
	Amount        float64   `json:"amount"`
	FileURL       string    `json:"file_url"`
	CreatedBy     uint64    `json:"created_by"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
