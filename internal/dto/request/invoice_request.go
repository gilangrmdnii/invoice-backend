package request

type CreateInvoiceRequest struct {
	ProjectID uint64  `json:"project_id" validate:"required"`
	Amount    float64 `json:"amount" validate:"required,gt=0"`
	FileURL   string  `json:"file_url" validate:"required,max=500"`
}

type UpdateInvoiceRequest struct {
	Amount  float64 `json:"amount" validate:"omitempty,gt=0"`
	FileURL string  `json:"file_url" validate:"omitempty,max=500"`
}
