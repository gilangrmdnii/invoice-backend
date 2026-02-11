package model

import "time"

type ProjectMember struct {
	ID        uint64    `json:"id"`
	ProjectID uint64    `json:"project_id"`
	UserID    uint64    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}
