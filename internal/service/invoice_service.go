package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

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

func (s *InvoiceService) notifyUser(ctx context.Context, userID uint64, title, message string, notifType model.NotificationType, refID uint64) {
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

func (s *InvoiceService) notifyRoles(ctx context.Context, roles []string, title, message string, notifType model.NotificationType, refID uint64) {
	users, err := s.userRepo.FindByRoles(ctx, roles)
	if err != nil {
		log.Printf("find users by roles error: %v", err)
		return
	}
	for _, u := range users {
		s.notifyUser(ctx, u.ID, title, message, notifType, refID)
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

	// Parse invoice date
	invoiceDate, err := time.Parse("2006-01-02", req.InvoiceDate)
	if err != nil {
		return nil, fmt.Errorf("invalid invoice date format, use YYYY-MM-DD")
	}

	// Calculate totals from items
	var subtotal float64
	items := make([]model.InvoiceItem, len(req.Items))
	for i, item := range req.Items {
		itemSubtotal := item.Quantity * item.UnitPrice
		subtotal += itemSubtotal
		items[i] = model.InvoiceItem{
			Description: item.Description,
			Quantity:    item.Quantity,
			Unit:        item.Unit,
			UnitPrice:   item.UnitPrice,
			Subtotal:    itemSubtotal,
		}
	}

	// Calculate tax and total
	taxAmount := subtotal * req.TaxPercentage / 100
	total := subtotal + taxAmount

	inv := &model.Invoice{
		InvoiceType:      model.InvoiceType(req.InvoiceType),
		ProjectID:        req.ProjectID,
		RecipientName:    req.RecipientName,
		RecipientAddress: req.RecipientAddress,
		Attention:        req.Attention,
		PONumber:         req.PONumber,
		InvoiceDate:      invoiceDate,
		DPPercentage:     req.DPPercentage,
		Subtotal:         subtotal,
		TaxPercentage:    req.TaxPercentage,
		TaxAmount:        taxAmount,
		Amount:           total,
		Notes:            req.Notes,
		Language:         req.Language,
		FileURL:          req.FileURL,
		CreatedBy:        userID,
	}

	id, err := s.invoiceRepo.Create(ctx, inv, items)
	if err != nil {
		return nil, fmt.Errorf("create invoice: %w", err)
	}

	s.logAudit(ctx, userID, "CREATE", id, fmt.Sprintf("type=%s, amount=%.2f", inv.InvoiceType, inv.Amount))
	s.notifyRoles(ctx, []string{"FINANCE", "OWNER"}, "New Invoice Created",
		fmt.Sprintf("A new %s invoice (%s) of %.2f has been created", inv.InvoiceType, inv.InvoiceNumber, inv.Amount),
		model.NotifInvoiceCreated, id)

	// Build response with items
	itemResponses := make([]response.InvoiceItemResponse, len(items))
	for i, item := range items {
		itemResponses[i] = response.InvoiceItemResponse{
			Description: item.Description,
			Quantity:    item.Quantity,
			Unit:        item.Unit,
			UnitPrice:   item.UnitPrice,
			Subtotal:    item.Subtotal,
			SortOrder:   i,
		}
	}

	return &response.InvoiceResponse{
		ID:               id,
		InvoiceNumber:    inv.InvoiceNumber,
		InvoiceType:      string(inv.InvoiceType),
		ProjectID:        inv.ProjectID,
		Amount:           inv.Amount,
		Status:           string(model.InvoiceStatusPending),
		FileURL:          inv.FileURL,
		RecipientName:    inv.RecipientName,
		RecipientAddress: inv.RecipientAddress,
		Attention:        inv.Attention,
		PONumber:         inv.PONumber,
		InvoiceDate:      inv.InvoiceDate.Format("2006-01-02"),
		DPPercentage:     inv.DPPercentage,
		Subtotal:         inv.Subtotal,
		TaxPercentage:    inv.TaxPercentage,
		TaxAmount:        inv.TaxAmount,
		Notes:            inv.Notes,
		Language:         inv.Language,
		CreatedBy:        userID,
		Items:            itemResponses,
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

	// Get items
	items, err := s.invoiceRepo.FindItemsByInvoiceID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get invoice items: %w", err)
	}

	resp := toInvoiceResponse(inv)
	resp.Items = make([]response.InvoiceItemResponse, len(items))
	for i, item := range items {
		resp.Items[i] = response.InvoiceItemResponse{
			ID:          item.ID,
			InvoiceID:   item.InvoiceID,
			Description: item.Description,
			Quantity:    item.Quantity,
			Unit:        item.Unit,
			UnitPrice:   item.UnitPrice,
			Subtotal:    item.Subtotal,
			SortOrder:   item.SortOrder,
		}
	}

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

	if inv.Status != model.InvoiceStatusPending {
		return nil, fmt.Errorf("only pending invoices can be updated")
	}

	if req.RecipientName != "" {
		inv.RecipientName = req.RecipientName
	}
	if req.RecipientAddress != "" {
		inv.RecipientAddress = req.RecipientAddress
	}
	if req.Attention != "" {
		inv.Attention = req.Attention
	}
	if req.PONumber != "" {
		inv.PONumber = req.PONumber
	}
	if req.InvoiceDate != "" {
		date, err := time.Parse("2006-01-02", req.InvoiceDate)
		if err != nil {
			return nil, fmt.Errorf("invalid invoice date format")
		}
		inv.InvoiceDate = date
	}
	if req.DPPercentage != nil {
		inv.DPPercentage = req.DPPercentage
	}
	if req.Notes != "" {
		inv.Notes = req.Notes
	}
	if req.Language != "" {
		inv.Language = req.Language
	}
	if req.FileURL != "" {
		inv.FileURL = req.FileURL
	}

	var items []model.InvoiceItem
	if req.Items != nil && len(req.Items) > 0 {
		var subtotal float64
		items = make([]model.InvoiceItem, len(req.Items))
		for i, item := range req.Items {
			itemSubtotal := item.Quantity * item.UnitPrice
			subtotal += itemSubtotal
			items[i] = model.InvoiceItem{
				Description: item.Description,
				Quantity:    item.Quantity,
				Unit:        item.Unit,
				UnitPrice:   item.UnitPrice,
				Subtotal:    itemSubtotal,
			}
		}
		inv.Subtotal = subtotal

		taxPct := inv.TaxPercentage
		if req.TaxPercentage != nil {
			taxPct = *req.TaxPercentage
		}
		inv.TaxPercentage = taxPct
		inv.TaxAmount = subtotal * taxPct / 100
		inv.Amount = subtotal + inv.TaxAmount
	} else if req.TaxPercentage != nil {
		inv.TaxPercentage = *req.TaxPercentage
		inv.TaxAmount = inv.Subtotal * *req.TaxPercentage / 100
		inv.Amount = inv.Subtotal + inv.TaxAmount
	}

	if err := s.invoiceRepo.Update(ctx, inv, items); err != nil {
		return nil, fmt.Errorf("update invoice: %w", err)
	}

	s.logAudit(ctx, userID, "UPDATE", id, "")

	return s.GetByID(ctx, id)
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

	if inv.Status != model.InvoiceStatusPending {
		return fmt.Errorf("only pending invoices can be deleted")
	}

	if err := s.invoiceRepo.Delete(ctx, id); err != nil {
		return err
	}

	s.logAudit(ctx, userID, "DELETE", id, "")
	return nil
}

func (s *InvoiceService) Approve(ctx context.Context, id uint64, approvedBy uint64, notes string) (*response.InvoiceResponse, error) {
	inv, err := s.invoiceRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("invoice not found")
		}
		return nil, err
	}

	if err := s.invoiceRepo.ApproveInvoice(ctx, id, approvedBy, notes); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("invoice not found")
		}
		if err.Error() == "invoice is not pending" {
			return nil, err
		}
		return nil, fmt.Errorf("approve invoice: %w", err)
	}

	s.logAudit(ctx, approvedBy, "APPROVE", id, notes)
	s.notifyUser(ctx, inv.CreatedBy, "Invoice Approved",
		fmt.Sprintf("Your invoice %s has been approved", inv.InvoiceNumber),
		model.NotifInvoiceApproved, id)

	return s.GetByID(ctx, id)
}

func (s *InvoiceService) Reject(ctx context.Context, id uint64, rejectedBy uint64, notes string) (*response.InvoiceResponse, error) {
	inv, err := s.invoiceRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("invoice not found")
		}
		return nil, err
	}

	if err := s.invoiceRepo.RejectInvoice(ctx, id, rejectedBy, notes); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("invoice not found")
		}
		if err.Error() == "invoice is not pending" {
			return nil, err
		}
		return nil, fmt.Errorf("reject invoice: %w", err)
	}

	s.logAudit(ctx, rejectedBy, "REJECT", id, notes)
	s.notifyUser(ctx, inv.CreatedBy, "Invoice Rejected",
		fmt.Sprintf("Your invoice %s has been rejected. Reason: %s", inv.InvoiceNumber, notes),
		model.NotifInvoiceRejected, id)

	return s.GetByID(ctx, id)
}

func toInvoiceResponse(inv *model.Invoice) response.InvoiceResponse {
	return response.InvoiceResponse{
		ID:               inv.ID,
		InvoiceNumber:    inv.InvoiceNumber,
		InvoiceType:      string(inv.InvoiceType),
		ProjectID:        inv.ProjectID,
		Amount:           inv.Amount,
		Status:           string(inv.Status),
		FileURL:          inv.FileURL,
		RecipientName:    inv.RecipientName,
		RecipientAddress: inv.RecipientAddress,
		Attention:        inv.Attention,
		PONumber:         inv.PONumber,
		InvoiceDate:      inv.InvoiceDate.Format("2006-01-02"),
		DPPercentage:     inv.DPPercentage,
		Subtotal:         inv.Subtotal,
		TaxPercentage:    inv.TaxPercentage,
		TaxAmount:        inv.TaxAmount,
		Notes:            inv.Notes,
		Language:         inv.Language,
		CreatedBy:        inv.CreatedBy,
		ApprovedBy:       inv.ApprovedBy,
		RejectNotes:      inv.RejectNotes,
		CreatedAt:        inv.CreatedAt,
		UpdatedAt:        inv.UpdatedAt,
	}
}
