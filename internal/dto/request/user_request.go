package request

type CreateUserRequest struct {
	FullName string `json:"full_name" validate:"required,min=2,max=255"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	Role     string `json:"role" validate:"required,oneof=SPV FINANCE OWNER"`
}

type UpdateUserRequest struct {
	FullName string `json:"full_name" validate:"omitempty,min=2,max=255"`
	Email    string `json:"email" validate:"omitempty,email"`
	Password string `json:"password" validate:"omitempty,min=6"`
	Role     string `json:"role" validate:"omitempty,oneof=SPV FINANCE OWNER"`
}
