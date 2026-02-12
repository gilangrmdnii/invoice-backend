package model

import "time"

type NotificationType string

const (
	NotifExpenseCreated  NotificationType = "EXPENSE_CREATED"
	NotifExpenseApproved NotificationType = "EXPENSE_APPROVED"
	NotifExpenseRejected NotificationType = "EXPENSE_REJECTED"
	NotifBudgetRequest   NotificationType = "BUDGET_REQUEST"
	NotifBudgetApproved  NotificationType = "BUDGET_APPROVED"
	NotifBudgetRejected  NotificationType = "BUDGET_REJECTED"
	NotifInvoiceCreated  NotificationType = "INVOICE_CREATED"
)

type Notification struct {
	ID          uint64           `json:"id"`
	UserID      uint64           `json:"user_id"`
	Title       string           `json:"title"`
	Message     string           `json:"message"`
	IsRead      bool             `json:"is_read"`
	Type        NotificationType `json:"type"`
	ReferenceID *uint64          `json:"reference_id,omitempty"`
	CreatedAt   time.Time        `json:"created_at"`
}
