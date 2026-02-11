package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	"github.com/gilangrmdnii/invoice-backend/internal/middleware"
	"github.com/gilangrmdnii/invoice-backend/internal/service"
	"github.com/gilangrmdnii/invoice-backend/pkg/response"
)

type NotificationHandler struct {
	notifService *service.NotificationService
}

func NewNotificationHandler(notifService *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{notifService: notifService}
}

func (h *NotificationHandler) List(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)

	notifications, err := h.notifService.ListByUser(c.Context(), userID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "failed to list notifications")
	}

	return response.Success(c, fiber.StatusOK, "notifications retrieved successfully", notifications)
}

func (h *NotificationHandler) CountUnread(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)

	result, err := h.notifService.CountUnread(c.Context(), userID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "failed to count unread notifications")
	}

	return response.Success(c, fiber.StatusOK, "unread count retrieved successfully", result)
}

func (h *NotificationHandler) MarkAsRead(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid notification id")
	}

	userID := middleware.GetUserID(c)

	if err := h.notifService.MarkAsRead(c.Context(), id, userID); err != nil {
		if err.Error() == "notification not found" {
			return response.Error(c, fiber.StatusNotFound, err.Error())
		}
		return response.Error(c, fiber.StatusInternalServerError, "failed to mark notification as read")
	}

	return response.Success(c, fiber.StatusOK, "notification marked as read", nil)
}

func (h *NotificationHandler) MarkAllAsRead(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)

	if err := h.notifService.MarkAllAsRead(c.Context(), userID); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "failed to mark all notifications as read")
	}

	return response.Success(c, fiber.StatusOK, "all notifications marked as read", nil)
}
