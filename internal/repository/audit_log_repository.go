package repository

import (
	"context"
	"database/sql"
	"encoding/json"

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
		// MySQL JSON column requires valid JSON; encode the string as a JSON string value
		b, _ := json.Marshal(log.Details)
		details = string(b)
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
	query := `SELECT a.id, a.user_id, COALESCE(u.full_name, ''), a.action, a.entity_type, a.entity_id, a.details, a.created_at
	FROM audit_logs a LEFT JOIN users u ON a.user_id = u.id ORDER BY a.created_at DESC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanAuditLogs(rows)
}

func (r *AuditLogRepository) FindByEntityType(ctx context.Context, entityType string) ([]model.AuditLog, error) {
	query := `SELECT a.id, a.user_id, COALESCE(u.full_name, ''), a.action, a.entity_type, a.entity_id, a.details, a.created_at
	FROM audit_logs a LEFT JOIN users u ON a.user_id = u.id WHERE a.entity_type = ? ORDER BY a.created_at DESC`
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
		if err := rows.Scan(&l.ID, &l.UserID, &l.FullName, &l.Action, &l.EntityType, &l.EntityID, &details, &l.CreatedAt); err != nil {
			return nil, err
		}
		if details.Valid {
			// MySQL JSON column wraps strings in quotes; try to unwrap
			var s string
			if err := json.Unmarshal([]byte(details.String), &s); err == nil {
				l.Details = s
			} else {
				l.Details = details.String
			}
		}
		logs = append(logs, l)
	}
	return logs, rows.Err()
}
