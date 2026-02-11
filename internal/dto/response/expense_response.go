package response

import "time"

type ExpenseResponse struct {
	ID          uint64    `json:"id"`
	ProjectID   uint64    `json:"project_id"`
	Description string    `json:"description"`
	Amount      float64   `json:"amount"`
	Category    string    `json:"category"`
	ReceiptURL  string    `json:"receipt_url,omitempty"`
	Status      string    `json:"status"`
	CreatedBy   uint64    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ExpenseApprovalResponse struct {
	ID         uint64    `json:"id"`
	ExpenseID  uint64    `json:"expense_id"`
	ApprovedBy uint64    `json:"approved_by"`
	Status     string    `json:"status"`
	Notes      string    `json:"notes,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}
