package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/gilangrmdnii/invoice-backend/internal/dto/request"
	"github.com/gilangrmdnii/invoice-backend/internal/dto/response"
	"github.com/gilangrmdnii/invoice-backend/internal/model"
	"github.com/gilangrmdnii/invoice-backend/internal/repository"
	"github.com/gilangrmdnii/invoice-backend/internal/sse"
)

type BudgetRequestService struct {
	budgetRequestRepo *repository.BudgetRequestRepository
	projectRepo       *repository.ProjectRepository
	memberRepo        *repository.ProjectMemberRepository
	budgetRepo        *repository.BudgetRepository
	auditRepo         *repository.AuditLogRepository
	notifRepo         *repository.NotificationRepository
	userRepo          *repository.UserRepository
	sseHub            *sse.Hub
}

func NewBudgetRequestService(
	budgetRequestRepo *repository.BudgetRequestRepository,
	projectRepo *repository.ProjectRepository,
	memberRepo *repository.ProjectMemberRepository,
	budgetRepo *repository.BudgetRepository,
	auditRepo *repository.AuditLogRepository,
	notifRepo *repository.NotificationRepository,
	userRepo *repository.UserRepository,
	sseHub *sse.Hub,
) *BudgetRequestService {
	return &BudgetRequestService{
		budgetRequestRepo: budgetRequestRepo,
		projectRepo:       projectRepo,
		memberRepo:        memberRepo,
		budgetRepo:        budgetRepo,
		auditRepo:         auditRepo,
		notifRepo:         notifRepo,
		userRepo:          userRepo,
		sseHub:            sseHub,
	}
}

func (s *BudgetRequestService) logAudit(ctx context.Context, userID uint64, action, entityType string, entityID uint64, details string) {
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

func (s *BudgetRequestService) notifyUser(ctx context.Context, userID uint64, title, message string, notifType model.NotificationType, refID uint64) {
	id, err := s.notifRepo.Create(ctx, &model.Notification{
		UserID:      userID,
		Title:       title,
		Message:     message,
		Type:        notifType,
		ReferenceID: &refID,
	})
	if err != nil {
		log.Printf("notification error: %v", err)
		return
	}
	s.sseHub.Publish(userID, sse.Event{
		Type: string(notifType),
		Data: map[string]interface{}{"id": id, "title": title, "message": message, "reference_id": refID},
	})
}

func (s *BudgetRequestService) notifyRoles(ctx context.Context, roles []string, title, message string, notifType model.NotificationType, refID uint64) {
	users, err := s.userRepo.FindByRoles(ctx, roles)
	if err != nil {
		log.Printf("find users by roles error: %v", err)
		return
	}
	for _, u := range users {
		s.notifyUser(ctx, u.ID, title, message, notifType, refID)
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

	proofURL := req.ProofURL
	br := &model.BudgetRequest{
		ProjectID:   req.ProjectID,
		RequestedBy: userID,
		Amount:      req.Amount,
		Reason:      req.Reason,
		ProofURL:    &proofURL,
		Status:      model.BudgetRequestPending,
	}

	id, err := s.budgetRequestRepo.Create(ctx, br)
	if err != nil {
		return nil, fmt.Errorf("create budget request: %w", err)
	}

	// Audit + Notification
	s.logAudit(ctx, userID, "CREATE", "budget_request", id, fmt.Sprintf("amount=%.2f, reason=%s", br.Amount, br.Reason))
	s.notifyRoles(ctx, []string{"FINANCE", "OWNER"}, "New Budget Request",
		fmt.Sprintf("A budget request of %.2f has been submitted", br.Amount),
		model.NotifBudgetRequest, id)

	return &response.BudgetRequestResponse{
		ID:          id,
		ProjectID:   br.ProjectID,
		RequestedBy: br.RequestedBy,
		Amount:      br.Amount,
		Reason:      br.Reason,
		ProofURL:    br.ProofURL,
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

func (s *BudgetRequestService) Approve(ctx context.Context, id, approvedBy uint64, notes, proofURL string) (*response.BudgetRequestResponse, error) {
	// Get before approve to know requester
	br, err := s.budgetRequestRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("budget request not found")
		}
		return nil, err
	}

	if err := s.budgetRequestRepo.ApproveBudgetRequest(ctx, id, approvedBy, notes, proofURL); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("budget request not found")
		}
		if err.Error() == "budget request is not pending" {
			return nil, err
		}
		return nil, fmt.Errorf("approve budget request: %w", err)
	}

	// Increase project total budget by approved amount
	if err := s.budgetRepo.IncreaseTotalBudget(ctx, br.ProjectID, br.Amount); err != nil {
		log.Printf("failed to increase budget for project %d: %v", br.ProjectID, err)
	}

	// Audit + Notification
	s.logAudit(ctx, approvedBy, "APPROVE", "budget_request", id, fmt.Sprintf("amount=%.2f added to project budget", br.Amount))
	s.notifyUser(ctx, br.RequestedBy, "Budget Request Approved",
		fmt.Sprintf("Your budget request of %.2f has been approved", br.Amount),
		model.NotifBudgetApproved, id)

	updated, err := s.budgetRequestRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	resp := toBudgetRequestResponse(updated)
	return &resp, nil
}

func (s *BudgetRequestService) Reject(ctx context.Context, id, approvedBy uint64, notes, proofURL string) (*response.BudgetRequestResponse, error) {
	// Get before reject to know requester
	br, err := s.budgetRequestRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("budget request not found")
		}
		return nil, err
	}

	if err := s.budgetRequestRepo.RejectBudgetRequest(ctx, id, approvedBy, notes, proofURL); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("budget request not found")
		}
		if err.Error() == "budget request is not pending" {
			return nil, err
		}
		return nil, fmt.Errorf("reject budget request: %w", err)
	}

	// Audit + Notification
	s.logAudit(ctx, approvedBy, "REJECT", "budget_request", id, "")
	s.notifyUser(ctx, br.RequestedBy, "Budget Request Rejected",
		fmt.Sprintf("Your budget request of %.2f has been rejected", br.Amount),
		model.NotifBudgetRejected, id)

	updated, err := s.budgetRequestRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	resp := toBudgetRequestResponse(updated)
	return &resp, nil
}

func toBudgetRequestResponse(br *model.BudgetRequest) response.BudgetRequestResponse {
	return response.BudgetRequestResponse{
		ID:               br.ID,
		ProjectID:        br.ProjectID,
		RequestedBy:      br.RequestedBy,
		Amount:           br.Amount,
		Reason:           br.Reason,
		ProofURL:         br.ProofURL,
		Status:           string(br.Status),
		ApprovedBy:       br.ApprovedBy,
		ApprovalNotes:    br.ApprovalNotes,
		ApprovalProofURL: br.ApprovalProofURL,
		CreatedAt:        br.CreatedAt,
		UpdatedAt:        br.UpdatedAt,
	}
}
