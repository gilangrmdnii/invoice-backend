package model

import "time"

type ProjectPlanItem struct {
	ID          uint64            `json:"id"`
	ProjectID   uint64            `json:"project_id"`
	ParentID    *uint64           `json:"parent_id,omitempty"`
	IsLabel     bool              `json:"is_label"`
	Description string            `json:"description"`
	Quantity    float64           `json:"quantity"`
	Unit        string            `json:"unit"`
	UnitPrice   float64           `json:"unit_price"`
	Subtotal    float64           `json:"subtotal"`
	SortOrder   int               `json:"sort_order"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	Children    []ProjectPlanItem `json:"-"` // transient, used during create
}
