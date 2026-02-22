package response

import "time"

type ProjectResponse struct {
	ID          uint64          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Status      string          `json:"status"`
	TotalBudget float64         `json:"total_budget"`
	SpentAmount float64         `json:"spent_amount"`
	CreatedBy   uint64          `json:"created_by"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type ProjectMemberResponse struct {
	ID        uint64    `json:"id"`
	ProjectID uint64    `json:"project_id"`
	UserID    uint64    `json:"user_id"`
	FullName  string    `json:"full_name"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}
