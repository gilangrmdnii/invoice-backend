package request

type CreateProjectWorkerRequest struct {
	ProjectID uint64  `json:"-"`
	FullName  string  `json:"full_name" validate:"required,min=2,max=255"`
	Role      string  `json:"role" validate:"required,min=2,max=100"`
	Phone     string  `json:"phone" validate:"omitempty,max=50"`
	DailyWage float64 `json:"daily_wage" validate:"omitempty,gte=0"`
}

type UpdateProjectWorkerRequest struct {
	FullName  string  `json:"full_name" validate:"omitempty,min=2,max=255"`
	Role      string  `json:"role" validate:"omitempty,min=2,max=100"`
	Phone     string  `json:"phone" validate:"omitempty,max=50"`
	DailyWage float64 `json:"daily_wage" validate:"omitempty,gte=0"`
	IsActive  *bool   `json:"is_active" validate:"omitempty"`
}
