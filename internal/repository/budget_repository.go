package repository

import (
	"context"
	"database/sql"

	"github.com/gilangrmdnii/invoice-backend/internal/model"
)

type BudgetRepository struct {
	db *sql.DB
}

func NewBudgetRepository(db *sql.DB) *BudgetRepository {
	return &BudgetRepository{db: db}
}

func (r *BudgetRepository) FindByProjectID(ctx context.Context, projectID uint64) (*model.ProjectBudget, error) {
	query := `SELECT id, project_id, total_budget, spent_amount, created_at, updated_at FROM project_budgets WHERE project_id = ?`
	b := &model.ProjectBudget{}
	err := r.db.QueryRowContext(ctx, query, projectID).Scan(
		&b.ID, &b.ProjectID, &b.TotalBudget, &b.SpentAmount, &b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (r *BudgetRepository) UpdateSpentAmount(ctx context.Context, projectID uint64, amount float64) error {
	query := `UPDATE project_budgets SET spent_amount = spent_amount + ? WHERE project_id = ?`
	_, err := r.db.ExecContext(ctx, query, amount, projectID)
	return err
}
