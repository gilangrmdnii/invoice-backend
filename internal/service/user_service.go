package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/gilangrmdnii/invoice-backend/internal/dto/request"
	"github.com/gilangrmdnii/invoice-backend/internal/dto/response"
	"github.com/gilangrmdnii/invoice-backend/internal/model"
	"github.com/gilangrmdnii/invoice-backend/internal/repository"
)

type UserService struct {
	userRepo  *repository.UserRepository
	auditRepo *repository.AuditLogRepository
}

func NewUserService(userRepo *repository.UserRepository, auditRepo *repository.AuditLogRepository) *UserService {
	return &UserService{userRepo: userRepo, auditRepo: auditRepo}
}

func (s *UserService) List(ctx context.Context, roleFilter string) ([]response.UserResponse, error) {
	var users []model.User
	var err error

	if roleFilter != "" {
		users, err = s.userRepo.FindByRoles(ctx, []string{roleFilter})
	} else {
		users, err = s.userRepo.FindAll(ctx)
	}
	if err != nil {
		return nil, err
	}

	result := make([]response.UserResponse, 0, len(users))
	for _, u := range users {
		result = append(result, toUserResponse(&u))
	}
	return result, nil
}

func (s *UserService) Create(ctx context.Context, req *request.CreateUserRequest, actorID uint64) (*response.UserResponse, error) {
	// Check email uniqueness
	existing, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("check email: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("email already registered")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := &model.User{
		FullName: req.FullName,
		Email:    req.Email,
		Password: string(hashed),
		Role:     model.UserRole(req.Role),
	}

	id, err := s.userRepo.Create(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	s.logAudit(ctx, actorID, "CREATE", "user", id, fmt.Sprintf("created user: %s (%s)", req.Email, req.Role))

	created, _ := s.userRepo.FindByID(ctx, id)
	if created != nil {
		resp := toUserResponse(created)
		return &resp, nil
	}

	return &response.UserResponse{
		ID:       id,
		FullName: user.FullName,
		Email:    user.Email,
		Role:     string(user.Role),
	}, nil
}

func (s *UserService) Update(ctx context.Context, targetID uint64, req *request.UpdateUserRequest, actorID uint64) (*response.UserResponse, error) {
	user, err := s.userRepo.FindByID(ctx, targetID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}

	// Cannot change own role
	if actorID == targetID && req.Role != "" && req.Role != string(user.Role) {
		return nil, fmt.Errorf("cannot change your own role")
	}

	// Check email uniqueness if changed
	if req.Email != "" && req.Email != user.Email {
		existing, err := s.userRepo.FindByEmail(ctx, req.Email)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("check email: %w", err)
		}
		if existing != nil {
			return nil, fmt.Errorf("email already registered")
		}
		user.Email = req.Email
	}

	if req.FullName != "" {
		user.FullName = req.FullName
	}
	if req.Role != "" {
		user.Role = model.UserRole(req.Role)
	}
	if req.Password != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("hash password: %w", err)
		}
		user.Password = string(hashed)
	}

	if err := s.userRepo.Update(ctx, targetID, user); err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}

	// Build change details
	var changes []string
	if req.FullName != "" {
		changes = append(changes, "name")
	}
	if req.Email != "" {
		changes = append(changes, "email")
	}
	if req.Role != "" {
		changes = append(changes, "role="+req.Role)
	}
	if req.Password != "" {
		changes = append(changes, "password")
	}
	s.logAudit(ctx, actorID, "UPDATE", "user", targetID, fmt.Sprintf("updated user: %s (%s)", user.Email, strings.Join(changes, ", ")))

	resp := toUserResponse(user)
	return &resp, nil
}

func (s *UserService) Delete(ctx context.Context, targetID, actorID uint64) error {
	if actorID == targetID {
		return fmt.Errorf("cannot delete yourself")
	}

	user, err := s.userRepo.FindByID(ctx, targetID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("user not found")
		}
		return err
	}

	if err := s.userRepo.Delete(ctx, targetID); err != nil {
		// Check for FK constraint error
		if strings.Contains(err.Error(), "foreign key constraint") || strings.Contains(err.Error(), "a]foreign key") {
			return fmt.Errorf("cannot delete user with associated data")
		}
		return fmt.Errorf("delete user: %w", err)
	}

	s.logAudit(ctx, actorID, "DELETE", "user", targetID, fmt.Sprintf("deleted user: %s (%s)", user.Email, string(user.Role)))
	return nil
}

func (s *UserService) logAudit(ctx context.Context, userID uint64, action, entityType string, entityID uint64, details string) {
	_, err := s.auditRepo.Create(ctx, &model.AuditLog{
		UserID:     userID,
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
		Details:    details,
	})
	if err != nil {
		log.Printf("audit log error: %v", err)
	}
}

func toUserResponse(u *model.User) response.UserResponse {
	return response.UserResponse{
		ID:        u.ID,
		FullName:  u.FullName,
		Email:     u.Email,
		Role:      string(u.Role),
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
