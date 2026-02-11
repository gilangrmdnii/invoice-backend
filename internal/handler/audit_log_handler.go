package handler

import (
	"github.com/gofiber/fiber/v2"

	"github.com/gilangrmdnii/invoice-backend/internal/dto/response"
	"github.com/gilangrmdnii/invoice-backend/internal/model"
	"github.com/gilangrmdnii/invoice-backend/internal/repository"
	pkgresponse "github.com/gilangrmdnii/invoice-backend/pkg/response"
)

type AuditLogHandler struct {
	auditRepo *repository.AuditLogRepository
}

func NewAuditLogHandler(auditRepo *repository.AuditLogRepository) *AuditLogHandler {
	return &AuditLogHandler{auditRepo: auditRepo}
}

func (h *AuditLogHandler) List(c *fiber.Ctx) error {
	entityType := c.Query("entity_type")

	var (
		logs []model.AuditLog
		err  error
	)

	if entityType != "" {
		logs, err = h.auditRepo.FindByEntityType(c.Context(), entityType)
	} else {
		logs, err = h.auditRepo.FindAll(c.Context())
	}

	if err != nil {
		return pkgresponse.Error(c, fiber.StatusInternalServerError, "failed to list audit logs")
	}

	result := make([]response.AuditLogResponse, 0, len(logs))
	for _, l := range logs {
		result = append(result, response.AuditLogResponse{
			ID:         l.ID,
			UserID:     l.UserID,
			Action:     l.Action,
			EntityType: l.EntityType,
			EntityID:   l.EntityID,
			Details:    l.Details,
			CreatedAt:  l.CreatedAt,
		})
	}

	return pkgresponse.Success(c, fiber.StatusOK, "audit logs retrieved successfully", result)
}
