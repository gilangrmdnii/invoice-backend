package request

type CreateBudgetRequestRequest struct {
	ProjectID uint64  `json:"project_id" validate:"required"`
	Amount    float64 `json:"amount" validate:"required,gt=0"`
	Reason    string  `json:"reason" validate:"required,min=2,max=1000"`
	ProofURL  string  `json:"proof_url" validate:"required"`
}

type ApproveBudgetRequestRequest struct {
	Notes    string `json:"notes"`
	ProofURL string `json:"proof_url" validate:"required"`
}

type RejectBudgetRequestRequest struct {
	Notes    string `json:"notes"`
	ProofURL string `json:"proof_url" validate:"required"`
}
