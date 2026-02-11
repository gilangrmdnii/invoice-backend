package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/gilangrmdnii/invoice-backend/internal/dto/request"
	"github.com/gilangrmdnii/invoice-backend/internal/dto/response"
	"github.com/gilangrmdnii/invoice-backend/internal/model"
	"github.com/gilangrmdnii/invoice-backend/internal/repository"
)

type ProjectService struct {
	projectRepo *repository.ProjectRepository
	memberRepo  *repository.ProjectMemberRepository
	budgetRepo  *repository.BudgetRepository
	userRepo    *repository.UserRepository
}

func NewProjectService(
	projectRepo *repository.ProjectRepository,
	memberRepo *repository.ProjectMemberRepository,
	budgetRepo *repository.BudgetRepository,
	userRepo *repository.UserRepository,
) *ProjectService {
	return &ProjectService{
		projectRepo: projectRepo,
		memberRepo:  memberRepo,
		budgetRepo:  budgetRepo,
		userRepo:    userRepo,
	}
}

func (s *ProjectService) Create(ctx context.Context, req *request.CreateProjectRequest, userID uint64) (*response.ProjectResponse, error) {
	project := &model.Project{
		Name:        req.Name,
		Description: req.Description,
		Status:      model.ProjectStatusActive,
		CreatedBy:   userID,
	}

	id, err := s.projectRepo.CreateWithBudget(ctx, project, req.TotalBudget)
	if err != nil {
		return nil, fmt.Errorf("create project: %w", err)
	}

	return &response.ProjectResponse{
		ID:          id,
		Name:        project.Name,
		Description: project.Description,
		Status:      string(project.Status),
		TotalBudget: req.TotalBudget,
		SpentAmount: 0,
		CreatedBy:   userID,
	}, nil
}

func (s *ProjectService) GetByID(ctx context.Context, id uint64) (*response.ProjectResponse, error) {
	project, err := s.projectRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("project not found")
		}
		return nil, err
	}

	budget, err := s.budgetRepo.FindByProjectID(ctx, id)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	resp := &response.ProjectResponse{
		ID:          project.ID,
		Name:        project.Name,
		Description: project.Description,
		Status:      string(project.Status),
		CreatedBy:   project.CreatedBy,
		CreatedAt:   project.CreatedAt,
		UpdatedAt:   project.UpdatedAt,
	}

	if budget != nil {
		resp.TotalBudget = budget.TotalBudget
		resp.SpentAmount = budget.SpentAmount
	}

	return resp, nil
}

func (s *ProjectService) List(ctx context.Context, userID uint64, role string) ([]response.ProjectResponse, error) {
	var projects []model.Project
	var err error

	if role == string(model.RoleSPV) {
		projects, err = s.projectRepo.FindByMemberUserID(ctx, userID)
	} else {
		projects, err = s.projectRepo.FindAll(ctx)
	}
	if err != nil {
		return nil, err
	}

	result := make([]response.ProjectResponse, 0, len(projects))
	for _, p := range projects {
		budget, _ := s.budgetRepo.FindByProjectID(ctx, p.ID)

		resp := response.ProjectResponse{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Status:      string(p.Status),
			CreatedBy:   p.CreatedBy,
			CreatedAt:   p.CreatedAt,
			UpdatedAt:   p.UpdatedAt,
		}

		if budget != nil {
			resp.TotalBudget = budget.TotalBudget
			resp.SpentAmount = budget.SpentAmount
		}

		result = append(result, resp)
	}

	return result, nil
}

func (s *ProjectService) Update(ctx context.Context, id uint64, req *request.UpdateProjectRequest) (*response.ProjectResponse, error) {
	project, err := s.projectRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("project not found")
		}
		return nil, err
	}

	if req.Name != "" {
		project.Name = req.Name
	}
	if req.Description != "" {
		project.Description = req.Description
	}
	if req.Status != "" {
		project.Status = model.ProjectStatus(req.Status)
	}

	if err := s.projectRepo.Update(ctx, project); err != nil {
		return nil, fmt.Errorf("update project: %w", err)
	}

	return s.GetByID(ctx, id)
}

func (s *ProjectService) AddMember(ctx context.Context, projectID, userID uint64) (*response.ProjectMemberResponse, error) {
	// Verify project exists
	_, err := s.projectRepo.FindByID(ctx, projectID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("project not found")
		}
		return nil, err
	}

	// Verify user exists and is SPV
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}
	if user.Role != model.RoleSPV {
		return nil, fmt.Errorf("only SPV users can be added as project members")
	}

	// Check if already a member
	exists, err := s.memberRepo.Exists(ctx, projectID, userID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("user is already a member of this project")
	}

	member := &model.ProjectMember{
		ProjectID: projectID,
		UserID:    userID,
	}

	id, err := s.memberRepo.Create(ctx, member)
	if err != nil {
		return nil, fmt.Errorf("add member: %w", err)
	}

	return &response.ProjectMemberResponse{
		ID:        id,
		ProjectID: projectID,
		UserID:    userID,
		FullName:  user.FullName,
		Email:     user.Email,
	}, nil
}

func (s *ProjectService) RemoveMember(ctx context.Context, projectID, userID uint64) error {
	if err := s.memberRepo.Delete(ctx, projectID, userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("member not found")
		}
		return err
	}
	return nil
}

func (s *ProjectService) ListMembers(ctx context.Context, projectID uint64) ([]response.ProjectMemberResponse, error) {
	// Verify project exists
	_, err := s.projectRepo.FindByID(ctx, projectID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("project not found")
		}
		return nil, err
	}

	members, err := s.memberRepo.FindByProjectID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	result := make([]response.ProjectMemberResponse, 0, len(members))
	for _, m := range members {
		user, err := s.userRepo.FindByID(ctx, m.UserID)
		if err != nil {
			continue
		}
		result = append(result, response.ProjectMemberResponse{
			ID:        m.ID,
			ProjectID: m.ProjectID,
			UserID:    m.UserID,
			FullName:  user.FullName,
			Email:     user.Email,
			CreatedAt: m.CreatedAt,
		})
	}

	return result, nil
}
