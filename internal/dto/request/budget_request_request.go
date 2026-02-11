package request

type CreateBudgetRequestRequest struct {
	ProjectID uint64  `json:"project_id" validate:"required"`
	Amount    float64 `json:"amount" validate:"required,gt=0"`
	Reason    string  `json:"reason" validate:"required,min=2,max=1000"`
}
