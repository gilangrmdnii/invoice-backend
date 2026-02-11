package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/gilangrmdnii/invoice-backend/internal/model"
)

type BudgetRequestRepository struct {
	db *sql.DB
}

func NewBudgetRequestRepository(db *sql.DB) *BudgetRequestRepository {
	return &BudgetRequestRepository{db: db}
}

func (r *BudgetRequestRepository) Create(ctx context.Context, br *model.BudgetRequest) (uint64, error) {
	query := `INSERT INTO budget_requests (project_id, requested_by, amount, reason, status) VALUES (?, ?, ?, ?, ?)`
	result, err := r.db.ExecContext(ctx, query,
		br.ProjectID, br.RequestedBy, br.Amount, br.Reason, br.Status,
	)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return uint64(id), nil
}

func (r *BudgetRequestRepository) FindByID(ctx context.Context, id uint64) (*model.BudgetRequest, error) {
	query := `SELECT id, project_id, requested_by, amount, reason, status, approved_by, created_at, updated_at FROM budget_requests WHERE id = ?`
	br := &model.BudgetRequest{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&br.ID, &br.ProjectID, &br.RequestedBy, &br.Amount, &br.Reason, &br.Status, &br.ApprovedBy, &br.CreatedAt, &br.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return br, nil
}

func (r *BudgetRequestRepository) FindAll(ctx context.Context) ([]model.BudgetRequest, error) {
	query := `SELECT id, project_id, requested_by, amount, reason, status, approved_by, created_at, updated_at FROM budget_requests ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []model.BudgetRequest
	for rows.Next() {
		var br model.BudgetRequest
		if err := rows.Scan(&br.ID, &br.ProjectID, &br.RequestedBy, &br.Amount, &br.Reason, &br.Status, &br.ApprovedBy, &br.CreatedAt, &br.UpdatedAt); err != nil {
			return nil, err
		}
		requests = append(requests, br)
	}
	return requests, rows.Err()
}

func (r *BudgetRequestRepository) FindByProjectIDs(ctx context.Context, projectIDs []uint64) ([]model.BudgetRequest, error) {
	if len(projectIDs) == 0 {
		return nil, nil
	}
	placeholders, args := buildInClause(projectIDs)
	query := fmt.Sprintf(`SELECT id, project_id, requested_by, amount, reason, status, approved_by, created_at, updated_at FROM budget_requests WHERE project_id IN (%s) ORDER BY created_at DESC`, placeholders)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []model.BudgetRequest
	for rows.Next() {
		var br model.BudgetRequest
		if err := rows.Scan(&br.ID, &br.ProjectID, &br.RequestedBy, &br.Amount, &br.Reason, &br.Status, &br.ApprovedBy, &br.CreatedAt, &br.UpdatedAt); err != nil {
			return nil, err
		}
		requests = append(requests, br)
	}
	return requests, rows.Err()
}

func (r *BudgetRequestRepository) ApproveBudgetRequest(ctx context.Context, id, approvedBy uint64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// Lock row and verify PENDING status
	var status model.BudgetRequestStatus
	var projectID uint64
	var amount float64
	err = tx.QueryRowContext(ctx,
		`SELECT status, project_id, amount FROM budget_requests WHERE id = ? FOR UPDATE`, id,
	).Scan(&status, &projectID, &amount)
	if err != nil {
		return err
	}
	if status != model.BudgetRequestPending {
		return fmt.Errorf("budget request is not pending")
	}

	// Update budget request status
	_, err = tx.ExecContext(ctx,
		`UPDATE budget_requests SET status = 'APPROVED', approved_by = ? WHERE id = ?`,
		approvedBy, id,
	)
	if err != nil {
		return fmt.Errorf("update budget request: %w", err)
	}

	// Update project budget total
	_, err = tx.ExecContext(ctx,
		`UPDATE project_budgets SET total_budget = total_budget + ? WHERE project_id = ?`,
		amount, projectID,
	)
	if err != nil {
		return fmt.Errorf("update budget: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

func (r *BudgetRequestRepository) RejectBudgetRequest(ctx context.Context, id, approvedBy uint64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// Lock row and verify PENDING status
	var status model.BudgetRequestStatus
	err = tx.QueryRowContext(ctx,
		`SELECT status FROM budget_requests WHERE id = ? FOR UPDATE`, id,
	).Scan(&status)
	if err != nil {
		return err
	}
	if status != model.BudgetRequestPending {
		return fmt.Errorf("budget request is not pending")
	}

	// Update budget request status (no budget change)
	_, err = tx.ExecContext(ctx,
		`UPDATE budget_requests SET status = 'REJECTED', approved_by = ? WHERE id = ?`,
		approvedBy, id,
	)
	if err != nil {
		return fmt.Errorf("update budget request: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}
