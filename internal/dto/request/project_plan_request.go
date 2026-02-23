package request

type PlanItemRequest struct {
	Description string  `json:"description" validate:"required,min=1,max=500"`
	Quantity    float64 `json:"quantity" validate:"required,gt=0"`
	Unit        string  `json:"unit" validate:"required,max=50"`
	UnitPrice   float64 `json:"unit_price" validate:"required,gt=0"`
}

type PlanLabelRequest struct {
	Description string            `json:"description" validate:"required,min=1,max=500"`
	Items       []PlanItemRequest `json:"items" validate:"required,min=1,dive"`
}

type UpdateProjectPlanRequest struct {
	Items  []PlanItemRequest  `json:"items" validate:"omitempty,dive"`
	Labels []PlanLabelRequest `json:"labels" validate:"omitempty,dive"`
}
