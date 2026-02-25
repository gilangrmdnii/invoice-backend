package request

type CreateProjectRequest struct {
	Name        string             `json:"name" validate:"required,min=2,max=255"`
	Description string             `json:"description" validate:"max=1000"`
	TotalBudget float64            `json:"total_budget" validate:"required,gt=0"`
	PlanItems   []PlanItemRequest  `json:"plan_items" validate:"omitempty,dive"`
	PlanLabels  []PlanLabelRequest `json:"plan_labels" validate:"omitempty,dive"`
}

type UpdateProjectRequest struct {
	Name        string `json:"name" validate:"omitempty,min=2,max=255"`
	Description string `json:"description" validate:"max=1000"`
	Status      string `json:"status" validate:"omitempty,oneof=ACTIVE COMPLETED ARCHIVED"`
}

type AddMemberRequest struct {
	UserID uint64 `json:"user_id" validate:"required"`
}
