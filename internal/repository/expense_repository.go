package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/gilangrmdnii/invoice-backend/internal/model"
)

type ExpenseRepository struct {
	db *sql.DB
}

func NewExpenseRepository(db *sql.DB) *ExpenseRepository {
	return &ExpenseRepository{db: db}
}

func (r *ExpenseRepository) Create(ctx context.Context, expense *model.Expense) (uint64, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	query := `INSERT INTO expenses (project_id, description, amount, category, receipt_url, created_by) VALUES (?, ?, ?, ?, ?, ?)`
	result, err := tx.ExecContext(ctx, query,
		expense.ProjectID, expense.Description, expense.Amount, expense.Category, expense.ReceiptURL, expense.CreatedBy,
	)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	// Update project budget spent amount
	_, err = tx.ExecContext(ctx,
		`UPDATE project_budgets SET spent_amount = spent_amount + ? WHERE project_id = ?`,
		expense.Amount, expense.ProjectID,
	)
	if err != nil {
		return 0, fmt.Errorf("update budget: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit tx: %w", err)
	}
	return uint64(id), nil
}

func (r *ExpenseRepository) FindByID(ctx context.Context, id uint64) (*model.Expense, error) {
	query := `SELECT id, project_id, description, amount, category, receipt_url, created_by, created_at, updated_at FROM expenses WHERE id = ?`
	e := &model.Expense{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&e.ID, &e.ProjectID, &e.Description, &e.Amount, &e.Category, &e.ReceiptURL, &e.CreatedBy, &e.CreatedAt, &e.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return e, nil
}

func (r *ExpenseRepository) FindAll(ctx context.Context) ([]model.Expense, error) {
	query := `SELECT id, project_id, description, amount, category, receipt_url, created_by, created_at, updated_at FROM expenses ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var expenses []model.Expense
	for rows.Next() {
		var e model.Expense
		if err := rows.Scan(&e.ID, &e.ProjectID, &e.Description, &e.Amount, &e.Category, &e.ReceiptURL, &e.CreatedBy, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, err
		}
		expenses = append(expenses, e)
	}
	return expenses, rows.Err()
}

func (r *ExpenseRepository) FindByProjectIDs(ctx context.Context, projectIDs []uint64) ([]model.Expense, error) {
	if len(projectIDs) == 0 {
		return nil, nil
	}
	placeholders, args := buildInClause(projectIDs)
	query := fmt.Sprintf(`SELECT id, project_id, description, amount, category, receipt_url, created_by, created_at, updated_at FROM expenses WHERE project_id IN (%s) ORDER BY created_at DESC`, placeholders)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var expenses []model.Expense
	for rows.Next() {
		var e model.Expense
		if err := rows.Scan(&e.ID, &e.ProjectID, &e.Description, &e.Amount, &e.Category, &e.ReceiptURL, &e.CreatedBy, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, err
		}
		expenses = append(expenses, e)
	}
	return expenses, rows.Err()
}

func (r *ExpenseRepository) Update(ctx context.Context, expense *model.Expense) error {
	query := `UPDATE expenses SET description = ?, amount = ?, category = ?, receipt_url = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, expense.Description, expense.Amount, expense.Category, expense.ReceiptURL, expense.ID)
	return err
}

func (r *ExpenseRepository) Delete(ctx context.Context, id uint64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// Get expense amount and project_id before deleting
	var amount float64
	var projectID uint64
	err = tx.QueryRowContext(ctx, `SELECT amount, project_id FROM expenses WHERE id = ? FOR UPDATE`, id).Scan(&amount, &projectID)
	if err != nil {
		if err == sql.ErrNoRows {
			return sql.ErrNoRows
		}
		return err
	}

	// Delete the expense
	result, err := tx.ExecContext(ctx, `DELETE FROM expenses WHERE id = ?`, id)
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

	// Deduct from project budget spent amount
	_, err = tx.ExecContext(ctx,
		`UPDATE project_budgets SET spent_amount = GREATEST(spent_amount - ?, 0) WHERE project_id = ?`,
		amount, projectID,
	)
	if err != nil {
		return fmt.Errorf("update budget: %w", err)
	}

	return tx.Commit()
}

