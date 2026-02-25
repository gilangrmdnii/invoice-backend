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
	query := `INSERT INTO expenses (project_id, description, amount, category, receipt_url, status, created_by) VALUES (?, ?, ?, ?, ?, ?, ?)`
	result, err := r.db.ExecContext(ctx, query,
		expense.ProjectID, expense.Description, expense.Amount, expense.Category, expense.ReceiptURL, expense.Status, expense.CreatedBy,
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

func (r *ExpenseRepository) FindByID(ctx context.Context, id uint64) (*model.Expense, error) {
	query := `SELECT id, project_id, description, amount, category, receipt_url, status, created_by, created_at, updated_at FROM expenses WHERE id = ?`
	e := &model.Expense{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&e.ID, &e.ProjectID, &e.Description, &e.Amount, &e.Category, &e.ReceiptURL, &e.Status, &e.CreatedBy, &e.CreatedAt, &e.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return e, nil
}

func (r *ExpenseRepository) FindAll(ctx context.Context) ([]model.Expense, error) {
	query := `SELECT id, project_id, description, amount, category, receipt_url, status, created_by, created_at, updated_at FROM expenses ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var expenses []model.Expense
	for rows.Next() {
		var e model.Expense
		if err := rows.Scan(&e.ID, &e.ProjectID, &e.Description, &e.Amount, &e.Category, &e.ReceiptURL, &e.Status, &e.CreatedBy, &e.CreatedAt, &e.UpdatedAt); err != nil {
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
	query := fmt.Sprintf(`SELECT id, project_id, description, amount, category, receipt_url, status, created_by, created_at, updated_at FROM expenses WHERE project_id IN (%s) ORDER BY created_at DESC`, placeholders)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var expenses []model.Expense
	for rows.Next() {
		var e model.Expense
		if err := rows.Scan(&e.ID, &e.ProjectID, &e.Description, &e.Amount, &e.Category, &e.ReceiptURL, &e.Status, &e.CreatedBy, &e.CreatedAt, &e.UpdatedAt); err != nil {
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
	query := `DELETE FROM expenses WHERE id = ?`
	result, err := r.db.ExecContext(ctx, query, id)
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

func (r *ExpenseRepository) ApproveExpense(ctx context.Context, expenseID, approvedBy uint64, notes string, proofURL string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// Lock row and verify PENDING status
	var status model.ExpenseStatus
	var projectID uint64
	var amount float64
	err = tx.QueryRowContext(ctx,
		`SELECT status, project_id, amount FROM expenses WHERE id = ? FOR UPDATE`, expenseID,
	).Scan(&status, &projectID, &amount)
	if err != nil {
		return err
	}
	if status != model.ExpenseStatusPending {
		return fmt.Errorf("expense is not pending")
	}

	// Update expense status
	_, err = tx.ExecContext(ctx,
		`UPDATE expenses SET status = 'APPROVED' WHERE id = ?`, expenseID,
	)
	if err != nil {
		return fmt.Errorf("update expense: %w", err)
	}

	// Insert approval record
	_, err = tx.ExecContext(ctx,
		`INSERT INTO expense_approvals (expense_id, approved_by, status, notes, proof_url) VALUES (?, ?, 'APPROVED', ?, ?)`,
		expenseID, approvedBy, notes, proofURL,
	)
	if err != nil {
		return fmt.Errorf("insert approval: %w", err)
	}

	// Update project budget spent amount
	_, err = tx.ExecContext(ctx,
		`UPDATE project_budgets SET spent_amount = spent_amount + ? WHERE project_id = ?`,
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

func (r *ExpenseRepository) RejectExpense(ctx context.Context, expenseID, approvedBy uint64, notes string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// Lock row and verify PENDING status
	var status model.ExpenseStatus
	err = tx.QueryRowContext(ctx,
		`SELECT status FROM expenses WHERE id = ? FOR UPDATE`, expenseID,
	).Scan(&status)
	if err != nil {
		return err
	}
	if status != model.ExpenseStatusPending {
		return fmt.Errorf("expense is not pending")
	}

	// Update expense status
	_, err = tx.ExecContext(ctx,
		`UPDATE expenses SET status = 'REJECTED' WHERE id = ?`, expenseID,
	)
	if err != nil {
		return fmt.Errorf("update expense: %w", err)
	}

	// Insert approval record
	_, err = tx.ExecContext(ctx,
		`INSERT INTO expense_approvals (expense_id, approved_by, status, notes) VALUES (?, ?, 'REJECTED', ?)`,
		expenseID, approvedBy, notes,
	)
	if err != nil {
		return fmt.Errorf("insert approval: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}
