package model

import "time"

type ProjectBudget struct {
	ID          uint64  `json:"id"`
	ProjectID   uint64  `json:"project_id"`
	TotalBudget float64 `json:"total_budget"`
	SpentAmount float64 `json:"spent_amount"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
