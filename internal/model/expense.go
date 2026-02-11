package model

import "time"

type ExpenseStatus string

const (
	ExpenseStatusPending  ExpenseStatus = "PENDING"
	ExpenseStatusApproved ExpenseStatus = "APPROVED"
	ExpenseStatusRejected ExpenseStatus = "REJECTED"
)

type Expense struct {
	ID          uint64        `json:"id"`
	ProjectID   uint64        `json:"project_id"`
	Description string        `json:"description"`
	Amount      float64       `json:"amount"`
	Category    string        `json:"category"`
	ReceiptURL  string        `json:"receipt_url,omitempty"`
	Status      ExpenseStatus `json:"status"`
	CreatedBy   uint64        `json:"created_by"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}
