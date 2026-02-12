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

type InvoiceService struct {
	invoiceRepo *repository.InvoiceRepository
	projectRepo *repository.ProjectRepository
	memberRepo  *repository.ProjectMemberRepository
	auditRepo   *repository.AuditLogRepository
	notifRepo   *repository.NotificationRepository
	userRepo    *repository.UserRepository
	sseHub      *sse.Hub
}

func NewInvoiceService(
	invoiceRepo *repository.InvoiceRepository,
	projectRepo *repository.ProjectRepository,
	memberRepo *repository.ProjectMemberRepository,
	auditRepo *repository.AuditLogRepository,
	notifRepo *repository.NotificationRepository,
	userRepo *repository.UserRepository,
	sseHub *sse.Hub,
) *InvoiceService {
	return &InvoiceService{
		invoiceRepo: invoiceRepo,
		projectRepo: projectRepo,
		memberRepo:  memberRepo,
		auditRepo:   auditRepo,
		notifRepo:   notifRepo,
		userRepo:    userRepo,
		sseHub:      sseHub,
	}
}

func (s *InvoiceService) logAudit(ctx context.Context, userID uint64, action string, entityID uint64, details string) {
	_, err := s.auditRepo.Create(ctx, &model.AuditLog{
		UserID:     userID,
		Action:     action,
		EntityType: "invoice",
		EntityID:   entityID,
		Details:    details,
	})
	if err != nil {
		log.Printf("audit log error: %v", err)
	}
}

func (s *InvoiceService) notifyRoles(ctx context.Context, roles []string, title, message string, notifType model.NotificationType, refID uint64) {
	users, err := s.userRepo.FindByRoles(ctx, roles)
	if err != nil {
		log.Printf("find users by roles error: %v", err)
		return
	}
	for _, u := range users {
		id, err := s.notifRepo.Create(ctx, &model.Notification{
			UserID:      u.ID,
			Title:       title,
			Message:     message,
			Type:        notifType,
			ReferenceID: &refID,
		})
		if err != nil {
			log.Printf("notification error: %v", err)
			continue
		}
		s.sseHub.Publish(u.ID, sse.Event{
			Type: string(notifType),
			Data: map[string]interface{}{"id": id, "title": title, "message": message, "reference_id": refID},
		})
	}
}

func (s *InvoiceService) Create(ctx context.Context, req *request.CreateInvoiceRequest, userID uint64) (*response.InvoiceResponse, error) {
	// Verify project exists
	_, err := s.projectRepo.FindByID(ctx, req.ProjectID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("project not found")
		}
		return nil, err
	}

	// SPV must be a member of the project
	isMember, err := s.memberRepo.Exists(ctx, req.ProjectID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, fmt.Errorf("not a member of this project")
	}

	inv := &model.Invoice{
		ProjectID: req.ProjectID,
		Amount:    req.Amount,
		FileURL:   req.FileURL,
		CreatedBy: userID,
	}

	id, err := s.invoiceRepo.Create(ctx, inv)
	if err != nil {
		return nil, fmt.Errorf("create invoice: %w", err)
	}

	s.logAudit(ctx, userID, "CREATE", id, fmt.Sprintf("amount=%.2f", inv.Amount))
	s.notifyRoles(ctx, []string{"FINANCE", "OWNER"}, "New Invoice Uploaded",
		fmt.Sprintf("A new invoice (%s) of %.2f has been uploaded", inv.InvoiceNumber, inv.Amount),
		model.NotifInvoiceCreated, id)

	return &response.InvoiceResponse{
		ID:            id,
		InvoiceNumber: inv.InvoiceNumber,
		ProjectID:     inv.ProjectID,
		Amount:        inv.Amount,
		FileURL:       inv.FileURL,
		CreatedBy:     userID,
	}, nil
}

func (s *InvoiceService) List(ctx context.Context, userID uint64, role string) ([]response.InvoiceResponse, error) {
	var invoices []model.Invoice
	var err error

	if role == string(model.RoleSPV) {
		projects, err := s.projectRepo.FindByMemberUserID(ctx, userID)
		if err != nil {
			return nil, err
		}
		projectIDs := make([]uint64, len(projects))
		for i, p := range projects {
			projectIDs[i] = p.ID
		}
		invoices, err = s.invoiceRepo.FindByProjectIDs(ctx, projectIDs)
		if err != nil {
			return nil, err
		}
	} else {
		invoices, err = s.invoiceRepo.FindAll(ctx)
	}
	if err != nil {
		return nil, err
	}

	result := make([]response.InvoiceResponse, 0, len(invoices))
	for _, inv := range invoices {
		result = append(result, toInvoiceResponse(&inv))
	}
	return result, nil
}

func (s *InvoiceService) GetByID(ctx context.Context, id uint64) (*response.InvoiceResponse, error) {
	inv, err := s.invoiceRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("invoice not found")
		}
		return nil, err
	}
	resp := toInvoiceResponse(inv)
	return &resp, nil
}

func (s *InvoiceService) Update(ctx context.Context, id uint64, req *request.UpdateInvoiceRequest, userID uint64) (*response.InvoiceResponse, error) {
	inv, err := s.invoiceRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("invoice not found")
		}
		return nil, err
	}

	if inv.CreatedBy != userID {
		return nil, fmt.Errorf("not authorized to update this invoice")
	}

	if req.Amount > 0 {
		inv.Amount = req.Amount
	}
	if req.FileURL != "" {
		inv.FileURL = req.FileURL
	}

	if err := s.invoiceRepo.Update(ctx, inv); err != nil {
		return nil, fmt.Errorf("update invoice: %w", err)
	}

	s.logAudit(ctx, userID, "UPDATE", id, "")

	updated, err := s.invoiceRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	resp := toInvoiceResponse(updated)
	return &resp, nil
}

func (s *InvoiceService) Delete(ctx context.Context, id uint64, userID uint64) error {
	inv, err := s.invoiceRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("invoice not found")
		}
		return err
	}

	if inv.CreatedBy != userID {
		return fmt.Errorf("not authorized to delete this invoice")
	}

	if err := s.invoiceRepo.Delete(ctx, id); err != nil {
		return err
	}

	s.logAudit(ctx, userID, "DELETE", id, "")
	return nil
}

func toInvoiceResponse(inv *model.Invoice) response.InvoiceResponse {
	return response.InvoiceResponse{
		ID:            inv.ID,
		InvoiceNumber: inv.InvoiceNumber,
		ProjectID:     inv.ProjectID,
		Amount:        inv.Amount,
		FileURL:       inv.FileURL,
		CreatedBy:     inv.CreatedBy,
		CreatedAt:     inv.CreatedAt,
		UpdatedAt:     inv.UpdatedAt,
	}
}
