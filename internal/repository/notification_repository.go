package repository

import (
	"context"
	"database/sql"

	"github.com/gilangrmdnii/invoice-backend/internal/model"
)

type NotificationRepository struct {
	db *sql.DB
}

func NewNotificationRepository(db *sql.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) Create(ctx context.Context, n *model.Notification) (uint64, error) {
	query := `INSERT INTO notifications (user_id, title, message, type, reference_id) VALUES (?, ?, ?, ?, ?)`
	result, err := r.db.ExecContext(ctx, query, n.UserID, n.Title, n.Message, n.Type, n.ReferenceID)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return uint64(id), nil
}

func (r *NotificationRepository) FindByUserID(ctx context.Context, userID uint64) ([]model.Notification, error) {
	query := `SELECT id, user_id, title, message, is_read, type, reference_id, created_at FROM notifications WHERE user_id = ? ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []model.Notification
	for rows.Next() {
		var n model.Notification
		if err := rows.Scan(&n.ID, &n.UserID, &n.Title, &n.Message, &n.IsRead, &n.Type, &n.ReferenceID, &n.CreatedAt); err != nil {
			return nil, err
		}
		notifications = append(notifications, n)
	}
	return notifications, rows.Err()
}

func (r *NotificationRepository) CountUnread(ctx context.Context, userID uint64) (int64, error) {
	query := `SELECT COUNT(1) FROM notifications WHERE user_id = ? AND is_read = false`
	var count int64
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&count)
	return count, err
}

func (r *NotificationRepository) MarkAsRead(ctx context.Context, id, userID uint64) error {
	query := `UPDATE notifications SET is_read = true WHERE id = ? AND user_id = ?`
	result, err := r.db.ExecContext(ctx, query, id, userID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *NotificationRepository) MarkAllAsRead(ctx context.Context, userID uint64) error {
	query := `UPDATE notifications SET is_read = true WHERE user_id = ? AND is_read = false`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}
