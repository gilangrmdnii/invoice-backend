package response

import "time"

type NotificationResponse struct {
	ID          uint64  `json:"id"`
	UserID      uint64  `json:"user_id"`
	Title       string  `json:"title"`
	Message     string  `json:"message"`
	IsRead      bool    `json:"is_read"`
	Type        string  `json:"type"`
	ReferenceID *uint64 `json:"reference_id,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type UnreadCountResponse struct {
	Count int64 `json:"count"`
}
