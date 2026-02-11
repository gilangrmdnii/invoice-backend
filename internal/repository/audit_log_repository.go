package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/gilangrmdnii/invoice-backend/internal/model"
)

type AuditLogRepository struct {
	db *sql.DB
}

func NewAuditLogRepository(db *sql.DB) *AuditLogRepository {
	return &AuditLogRepository{db: db}
}

func (r *AuditLogRepository) Create(ctx context.Context, log *model.AuditLog) (uint64, error) {
	query := `INSERT INTO audit_logs (user_id, action, entity_type, entity_id, details) VALUES (?, ?, ?, ?, ?)`
	var details interface{}
	if log.Details != "" {
		details = log.Details
	}
	result, err := r.db.ExecContext(ctx, query, log.UserID, log.Action, log.EntityType, log.EntityID, details)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return uint64(id), nil
}

func (r *AuditLogRepository) FindAll(ctx context.Context) ([]model.AuditLog, error) {
	query := `SELECT id, user_id, action, entity_type, entity_id, details, created_at FROM audit_logs ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanAuditLogs(rows)
}

func (r *AuditLogRepository) FindByEntityType(ctx context.Context, entityType string) ([]model.AuditLog, error) {
	query := fmt.Sprintf(`SELECT id, user_id, action, entity_type, entity_id, details, created_at FROM audit_logs WHERE entity_type = ? ORDER BY created_at DESC`)
	rows, err := r.db.QueryContext(ctx, query, entityType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanAuditLogs(rows)
}

func scanAuditLogs(rows *sql.Rows) ([]model.AuditLog, error) {
	var logs []model.AuditLog
	for rows.Next() {
		var l model.AuditLog
		var details sql.NullString
		if err := rows.Scan(&l.ID, &l.UserID, &l.Action, &l.EntityType, &l.EntityID, &details, &l.CreatedAt); err != nil {
			return nil, err
		}
		if details.Valid {
			l.Details = details.String
		}
		logs = append(logs, l)
	}
	return logs, rows.Err()
}
