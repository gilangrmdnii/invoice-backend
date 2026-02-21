package model

import "time"

type InvoiceItem struct {
	ID          uint64        `json:"id"`
	InvoiceID   uint64        `json:"invoice_id"`
	ParentID    *uint64       `json:"parent_id,omitempty"`
	IsLabel     bool          `json:"is_label"`
	Description string        `json:"description"`
	Quantity    float64       `json:"quantity"`
	Unit        string        `json:"unit"`
	UnitPrice   float64       `json:"unit_price"`
	Subtotal    float64       `json:"subtotal"`
	SortOrder   int           `json:"sort_order"`
	CreatedAt   time.Time     `json:"created_at"`
	Children    []InvoiceItem `json:"-"` // transient, used during create
}
