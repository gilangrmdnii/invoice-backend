package model

import "time"

type ProjectWorker struct {
	ID        uint64    `json:"id"`
	ProjectID uint64    `json:"project_id"`
	FullName  string    `json:"full_name"`
	Role      string    `json:"role"`
	Phone     string    `json:"phone"`
	DailyWage float64   `json:"daily_wage"`
	IsActive  bool      `json:"is_active"`
	AddedBy   uint64    `json:"added_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
