package model

import "time"

type UserRole string

const (
	RoleSPV     UserRole = "SPV"
	RoleFinance UserRole = "FINANCE"
	RoleOwner   UserRole = "OWNER"
)

type User struct {
	ID        uint64   `json:"id"`
	FullName  string   `json:"full_name"`
	Email     string   `json:"email"`
	Password  string   `json:"-"`
	Role      UserRole `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
