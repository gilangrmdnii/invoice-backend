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

type BudgetRequestService struct {
	budgetRequestRepo *repository.BudgetRequestRepository
	projectRepo       *repository.ProjectRepository
	memberRepo        *repository.ProjectMemberRepository
}

func NewBudgetRequestService(
	budgetRequestRepo *repository.BudgetRequestRepository,
	projectRepo *repository.ProjectRepository,
	memberRepo *repository.ProjectMemberRepository,
) *BudgetRequestService {
	return &BudgetRequestService{
		budgetRequestRepo: budgetRequestRepo,
		projectRepo:       projectRepo,
		memberRepo:        memberRepo,
	}
}

func (s *BudgetRequestService) Create(ctx context.Context, req *request.CreateBudgetRequestRequest, userID uint64, role string) (*response.BudgetRequestResponse, error) {
	// Verify project exists
	_, err := s.projectRepo.FindByID(ctx, req.ProjectID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("project not found")
		}
		return nil, err
	}

	// SPV must be a member of the project
	if role == string(model.RoleSPV) {
		isMember, err := s.memberRepo.Exists(ctx, req.ProjectID, userID)
		if err != nil {
			return nil, err
		}
		if !isMember {
			return nil, fmt.Errorf("not a member of this project")
		}
	}

	br := &model.BudgetRequest{
		ProjectID:   req.ProjectID,
		RequestedBy: userID,
		Amount:      req.Amount,
		Reason:      req.Reason,
		Status:      model.BudgetRequestPending,
	}

	id, err := s.budgetRequestRepo.Create(ctx, br)
	if err != nil {
		return nil, fmt.Errorf("create budget request: %w", err)
	}

	return &response.BudgetRequestResponse{
		ID:          id,
		ProjectID:   br.ProjectID,
		RequestedBy: br.RequestedBy,
		Amount:      br.Amount,
		Reason:      br.Reason,
		Status:      string(br.Status),
	}, nil
}

func (s *BudgetRequestService) List(ctx context.Context, userID uint64, role string) ([]response.BudgetRequestResponse, error) {
	var requests []model.BudgetRequest
	var err error

	if role == string(model.RoleSPV) {
		// SPV sees only budget requests from their projects
		projects, err := s.projectRepo.FindByMemberUserID(ctx, userID)
		if err != nil {
			return nil, err
		}
		projectIDs := make([]uint64, len(projects))
		for i, p := range projects {
			projectIDs[i] = p.ID
		}
		requests, err = s.budgetRequestRepo.FindByProjectIDs(ctx, projectIDs)
		if err != nil {
			return nil, err
		}
	} else {
		requests, err = s.budgetRequestRepo.FindAll(ctx)
	}
	if err != nil {
		return nil, err
	}

	result := make([]response.BudgetRequestResponse, 0, len(requests))
	for _, br := range requests {
		result = append(result, toBudgetRequestResponse(&br))
	}
	return result, nil
}

func (s *BudgetRequestService) GetByID(ctx context.Context, id uint64) (*response.BudgetRequestResponse, error) {
	br, err := s.budgetRequestRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("budget request not found")
		}
		return nil, err
	}
	resp := toBudgetRequestResponse(br)
	return &resp, nil
}

func (s *BudgetRequestService) Approve(ctx context.Context, id, approvedBy uint64) (*response.BudgetRequestResponse, error) {
	if err := s.budgetRequestRepo.ApproveBudgetRequest(ctx, id, approvedBy); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("budget request not found")
		}
		if err.Error() == "budget request is not pending" {
			return nil, err
		}
		return nil, fmt.Errorf("approve budget request: %w", err)
	}

	br, err := s.budgetRequestRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	resp := toBudgetRequestResponse(br)
	return &resp, nil
}

func (s *BudgetRequestService) Reject(ctx context.Context, id, approvedBy uint64) (*response.BudgetRequestResponse, error) {
	if err := s.budgetRequestRepo.RejectBudgetRequest(ctx, id, approvedBy); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("budget request not found")
		}
		if err.Error() == "budget request is not pending" {
			return nil, err
		}
		return nil, fmt.Errorf("reject budget request: %w", err)
	}

	br, err := s.budgetRequestRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	resp := toBudgetRequestResponse(br)
	return &resp, nil
}

func toBudgetRequestResponse(br *model.BudgetRequest) response.BudgetRequestResponse {
	return response.BudgetRequestResponse{
		ID:          br.ID,
		ProjectID:   br.ProjectID,
		RequestedBy: br.RequestedBy,
		Amount:      br.Amount,
		Reason:      br.Reason,
		Status:      string(br.Status),
		ApprovedBy:  br.ApprovedBy,
		CreatedAt:   br.CreatedAt,
		UpdatedAt:   br.UpdatedAt,
	}
}
