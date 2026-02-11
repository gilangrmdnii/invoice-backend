package model

import "time"

type BudgetRequestStatus string

const (
	BudgetRequestPending  BudgetRequestStatus = "PENDING"
	BudgetRequestApproved BudgetRequestStatus = "APPROVED"
	BudgetRequestRejected BudgetRequestStatus = "REJECTED"
)

type BudgetRequest struct {
	ID          uint64              `json:"id"`
	ProjectID   uint64              `json:"project_id"`
	RequestedBy uint64              `json:"requested_by"`
	Amount      float64             `json:"amount"`
	Reason      string              `json:"reason"`
	Status      BudgetRequestStatus `json:"status"`
	ApprovedBy  *uint64             `json:"approved_by,omitempty"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
}
