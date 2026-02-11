package model

import "time"

type ApprovalStatus string

const (
	ApprovalStatusApproved ApprovalStatus = "APPROVED"
	ApprovalStatusRejected ApprovalStatus = "REJECTED"
)

type ExpenseApproval struct {
	ID         uint64         `json:"id"`
	ExpenseID  uint64         `json:"expense_id"`
	ApprovedBy uint64         `json:"approved_by"`
	Status     ApprovalStatus `json:"status"`
	Notes      string         `json:"notes,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
}
