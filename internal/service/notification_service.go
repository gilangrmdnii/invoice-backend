package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/gilangrmdnii/invoice-backend/internal/dto/response"
	"github.com/gilangrmdnii/invoice-backend/internal/repository"
)

type NotificationService struct {
	notifRepo *repository.NotificationRepository
}

func NewNotificationService(notifRepo *repository.NotificationRepository) *NotificationService {
	return &NotificationService{notifRepo: notifRepo}
}

func (s *NotificationService) ListByUser(ctx context.Context, userID uint64) ([]response.NotificationResponse, error) {
	notifications, err := s.notifRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make([]response.NotificationResponse, 0, len(notifications))
	for _, n := range notifications {
		result = append(result, response.NotificationResponse{
			ID:          n.ID,
			UserID:      n.UserID,
			Title:       n.Title,
			Message:     n.Message,
			IsRead:      n.IsRead,
			Type:        string(n.Type),
			ReferenceID: n.ReferenceID,
			CreatedAt:   n.CreatedAt,
		})
	}
	return result, nil
}

func (s *NotificationService) CountUnread(ctx context.Context, userID uint64) (*response.UnreadCountResponse, error) {
	count, err := s.notifRepo.CountUnread(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &response.UnreadCountResponse{Count: count}, nil
}

func (s *NotificationService) MarkAsRead(ctx context.Context, id, userID uint64) error {
	err := s.notifRepo.MarkAsRead(ctx, id, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("notification not found")
		}
		return err
	}
	return nil
}

func (s *NotificationService) MarkAllAsRead(ctx context.Context, userID uint64) error {
	return s.notifRepo.MarkAllAsRead(ctx, userID)
}
