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

type InvoicePaymentService struct {
	paymentRepo *repository.InvoicePaymentRepository
	invoiceRepo *repository.InvoiceRepository
	auditRepo   *repository.AuditLogRepository
	notifRepo   *repository.NotificationRepository
	userRepo    *repository.UserRepository
	sseHub      *sse.Hub
}

func NewInvoicePaymentService(
	paymentRepo *repository.InvoicePaymentRepository,
	invoiceRepo *repository.InvoiceRepository,
	auditRepo *repository.AuditLogRepository,
	notifRepo *repository.NotificationRepository,
	userRepo *repository.UserRepository,
	sseHub *sse.Hub,
) *InvoicePaymentService {
	return &InvoicePaymentService{
		paymentRepo: paymentRepo,
		invoiceRepo: invoiceRepo,
		auditRepo:   auditRepo,
		notifRepo:   notifRepo,
		userRepo:    userRepo,
		sseHub:      sseHub,
	}
}

func (s *InvoicePaymentService) Create(ctx context.Context, req *request.CreateInvoicePaymentRequest, userID uint64) (*response.InvoicePaymentResponse, error) {
	// Verify invoice exists and is approved
	inv, err := s.invoiceRepo.FindByID(ctx, req.InvoiceID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("invoice not found")
		}
		return nil, err
	}

	if inv.Status != model.InvoiceStatusApproved {
		return nil, fmt.Errorf("invoice must be approved before recording payments")
	}

	if inv.PaymentStatus == model.PaymentStatusPaid {
		return nil, fmt.Errorf("invoice is already fully paid")
	}

	// Validate payment amount doesn't exceed remaining
	remaining := inv.Amount - inv.PaidAmount
	if req.Amount > remaining {
		return nil, fmt.Errorf("payment amount (%.2f) exceeds remaining balance (%.2f)", req.Amount, remaining)
	}

	paymentDate, err := time.Parse("2006-01-02", req.PaymentDate)
	if err != nil {
		return nil, fmt.Errorf("invalid payment date format, use YYYY-MM-DD")
	}

	payment := &model.InvoicePayment{
		InvoiceID:     req.InvoiceID,
		Amount:        req.Amount,
		PaymentDate:   paymentDate,
		PaymentMethod: model.PaymentMethod(req.PaymentMethod),
		ProofURL:      req.ProofURL,
		Notes:         req.Notes,
		CreatedBy:     userID,
	}

	id, err := s.paymentRepo.Create(ctx, payment)
	if err != nil {
		return nil, fmt.Errorf("create payment: %w", err)
	}

	// Audit log
	s.logAudit(ctx, userID, "CREATE", "invoice_payment", id,
		fmt.Sprintf("invoice=%s, amount=%.2f, method=%s", inv.InvoiceNumber, req.Amount, req.PaymentMethod))

	// Notify invoice creator
	newPaid := inv.PaidAmount + req.Amount
	statusLabel := "Partial Payment"
	if newPaid >= inv.Amount {
		statusLabel = "Fully Paid"
	}
	s.notifyUser(ctx, inv.CreatedBy, "Payment Recorded",
		fmt.Sprintf("Payment of %.2f recorded for invoice %s (%s)", req.Amount, inv.InvoiceNumber, statusLabel),
		model.NotifInvoiceApproved, inv.ID)

	// Get creator name
	creator, _ := s.userRepo.FindByID(ctx, userID)
	creatorName := ""
	if creator != nil {
		creatorName = creator.FullName
	}

	return &response.InvoicePaymentResponse{
		ID:            id,
		InvoiceID:     payment.InvoiceID,
		Amount:        payment.Amount,
		PaymentDate:   payment.PaymentDate.Format("2006-01-02"),
		PaymentMethod: string(payment.PaymentMethod),
		ProofURL:      payment.ProofURL,
		Notes:         payment.Notes,
		CreatedBy:     payment.CreatedBy,
		CreatorName:   creatorName,
		CreatedAt:     time.Now(),
	}, nil
}

func (s *InvoicePaymentService) ListByInvoice(ctx context.Context, invoiceID uint64) ([]response.InvoicePaymentResponse, error) {
	payments, err := s.paymentRepo.FindByInvoiceID(ctx, invoiceID)
	if err != nil {
		return nil, err
	}

	result := make([]response.InvoicePaymentResponse, 0, len(payments))
	for _, p := range payments {
		// Get creator name
		creator, _ := s.userRepo.FindByID(ctx, p.CreatedBy)
		creatorName := ""
		if creator != nil {
			creatorName = creator.FullName
		}

		result = append(result, response.InvoicePaymentResponse{
			ID:            p.ID,
			InvoiceID:     p.InvoiceID,
			Amount:        p.Amount,
			PaymentDate:   p.PaymentDate.Format("2006-01-02"),
			PaymentMethod: string(p.PaymentMethod),
			ProofURL:      p.ProofURL,
			Notes:         p.Notes,
			CreatedBy:     p.CreatedBy,
			CreatorName:   creatorName,
			CreatedAt:     p.CreatedAt,
		})
	}
	return result, nil
}

func (s *InvoicePaymentService) Delete(ctx context.Context, paymentID, invoiceID, userID uint64) error {
	if err := s.paymentRepo.Delete(ctx, paymentID, invoiceID); err != nil {
		return fmt.Errorf("delete payment: %w", err)
	}

	s.logAudit(ctx, userID, "DELETE", "invoice_payment", paymentID, fmt.Sprintf("invoice_id=%d", invoiceID))
	return nil
}

func (s *InvoicePaymentService) logAudit(ctx context.Context, userID uint64, action, entityType string, entityID uint64, details string) {
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

func (s *InvoicePaymentService) notifyUser(ctx context.Context, userID uint64, title, message string, notifType model.NotificationType, refID uint64) {
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
