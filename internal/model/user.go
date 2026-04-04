package model

import "time"

type UserRole string

const (
	RoleSPV     UserRole = "SPV"
	RoleFinance UserRole = "FINANCE"
	RoleOwner   UserRole = "OWNER"
	RoleQC      UserRole = "QC"
)

// IsFieldRole returns true for roles that are scoped to their own projects (SPV, QC)
func IsFieldRole(role string) bool {
	return role == string(RoleSPV) || role == string(RoleQC)
}

type User struct {
	ID        uint64   `json:"id"`
	FullName  string   `json:"full_name"`
	Email     string   `json:"email"`
	Password  string   `json:"-"`
	Role      UserRole `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
