package response

import "time"

type AuditLogResponse struct {
	ID         uint64    `json:"id"`
	UserID     uint64    `json:"user_id"`
	FullName   string    `json:"full_name"`
	Action     string    `json:"action"`
	EntityType string    `json:"entity_type"`
	EntityID   uint64    `json:"entity_id"`
	Details    string    `json:"details,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}
