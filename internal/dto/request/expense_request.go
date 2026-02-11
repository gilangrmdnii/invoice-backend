package request

type CreateExpenseRequest struct {
	ProjectID   uint64  `json:"project_id" validate:"required"`
	Description string  `json:"description" validate:"required,min=2,max=1000"`
	Amount      float64 `json:"amount" validate:"required,gt=0"`
	Category    string  `json:"category" validate:"required,max=255"`
	ReceiptURL  string  `json:"receipt_url" validate:"omitempty,url,max=500"`
}

type UpdateExpenseRequest struct {
	Description string  `json:"description" validate:"omitempty,min=2,max=1000"`
	Amount      float64 `json:"amount" validate:"omitempty,gt=0"`
	Category    string  `json:"category" validate:"omitempty,max=255"`
	ReceiptURL  string  `json:"receipt_url" validate:"omitempty,url,max=500"`
}

type ApproveExpenseRequest struct {
	Notes string `json:"notes" validate:"max=1000"`
}
