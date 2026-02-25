package repository

import (
	"context"
	"database/sql"
	"fmt"
)

type ProjectSummaryRow struct {
	TotalProjects  int64
	ActiveProjects int64
}

type BudgetSummaryRow struct {
	TotalBudget     float64
	TotalPlanBudget float64
	TotalSpent      float64
	Remaining       float64
}

type ExpenseSummaryRow struct {
	TotalExpenses int64
	TotalAmount   float64
}

type BudgetRequestSummaryRow struct {
	TotalRequests    int64
	PendingRequests  int64
	ApprovedRequests int64
	RejectedRequests int64
	TotalAmount      float64
}

type InvoiceSummaryRow struct {
	TotalInvoices int64
	TotalAmount   float64
}

type DashboardRepository struct {
	db *sql.DB
}

func NewDashboardRepository(db *sql.DB) *DashboardRepository {
	return &DashboardRepository{db: db}
}

func (r *DashboardRepository) GetProjectSummary(ctx context.Context, projectIDs []uint64) (*ProjectSummaryRow, error) {
	var query string
	var args []interface{}

	if len(projectIDs) > 0 {
		placeholders, pArgs := buildInClause(projectIDs)
		query = fmt.Sprintf(`SELECT COUNT(1), SUM(CASE WHEN status = 'ACTIVE' THEN 1 ELSE 0 END) FROM projects WHERE id IN (%s)`, placeholders)
		args = pArgs
	} else {
		query = `SELECT COUNT(1), SUM(CASE WHEN status = 'ACTIVE' THEN 1 ELSE 0 END) FROM projects`
	}

	row := &ProjectSummaryRow{}
	var active sql.NullInt64
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&row.TotalProjects, &active)
	if err != nil {
		return nil, err
	}
	row.ActiveProjects = active.Int64
	return row, nil
}

func (r *DashboardRepository) GetBudgetSummary(ctx context.Context, projectIDs []uint64) (*BudgetSummaryRow, error) {
	var query string
	var args []interface{}

	if len(projectIDs) > 0 {
		placeholders, pArgs := buildInClause(projectIDs)
		query = fmt.Sprintf(`SELECT COALESCE(SUM(pb.total_budget),0), COALESCE(SUM(pb.spent_amount),0),
			COALESCE((SELECT SUM(ppi.subtotal) FROM project_plan_items ppi
				INNER JOIN project_plans pp ON pp.id = ppi.plan_id
				WHERE pp.project_id IN (%s) AND ppi.is_label = false), 0)
			FROM project_budgets pb WHERE pb.project_id IN (%s)`, placeholders, placeholders)
		args = append(pArgs, pArgs...)
	} else {
		query = `SELECT COALESCE(SUM(pb.total_budget),0), COALESCE(SUM(pb.spent_amount),0),
			COALESCE((SELECT SUM(ppi.subtotal) FROM project_plan_items ppi
				INNER JOIN project_plans pp ON pp.id = ppi.plan_id
				WHERE ppi.is_label = false), 0)
			FROM project_budgets pb`
	}

	row := &BudgetSummaryRow{}
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&row.TotalBudget, &row.TotalSpent, &row.TotalPlanBudget)
	if err != nil {
		return nil, err
	}
	row.Remaining = row.TotalBudget - row.TotalSpent
	return row, nil
}

func (r *DashboardRepository) GetExpenseSummary(ctx context.Context, projectIDs []uint64) (*ExpenseSummaryRow, error) {
	var query string
	var args []interface{}

	if len(projectIDs) > 0 {
		placeholders, pArgs := buildInClause(projectIDs)
		query = fmt.Sprintf(`SELECT COUNT(1), COALESCE(SUM(amount),0) FROM expenses WHERE project_id IN (%s)`, placeholders)
		args = pArgs
	} else {
		query = `SELECT COUNT(1), COALESCE(SUM(amount),0) FROM expenses`
	}

	row := &ExpenseSummaryRow{}
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&row.TotalExpenses, &row.TotalAmount)
	if err != nil {
		return nil, err
	}
	return row, nil
}

func (r *DashboardRepository) GetBudgetRequestSummary(ctx context.Context, projectIDs []uint64) (*BudgetRequestSummaryRow, error) {
	var query string
	var args []interface{}

	if len(projectIDs) > 0 {
		placeholders, pArgs := buildInClause(projectIDs)
		query = fmt.Sprintf(`SELECT COUNT(1),
			SUM(CASE WHEN status = 'PENDING' THEN 1 ELSE 0 END),
			SUM(CASE WHEN status = 'APPROVED' THEN 1 ELSE 0 END),
			SUM(CASE WHEN status = 'REJECTED' THEN 1 ELSE 0 END),
			COALESCE(SUM(amount),0)
			FROM budget_requests WHERE project_id IN (%s)`, placeholders)
		args = pArgs
	} else {
		query = `SELECT COUNT(1),
			SUM(CASE WHEN status = 'PENDING' THEN 1 ELSE 0 END),
			SUM(CASE WHEN status = 'APPROVED' THEN 1 ELSE 0 END),
			SUM(CASE WHEN status = 'REJECTED' THEN 1 ELSE 0 END),
			COALESCE(SUM(amount),0)
			FROM budget_requests`
	}

	row := &BudgetRequestSummaryRow{}
	var pending, approved, rejected sql.NullInt64
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&row.TotalRequests, &pending, &approved, &rejected, &row.TotalAmount)
	if err != nil {
		return nil, err
	}
	row.PendingRequests = pending.Int64
	row.ApprovedRequests = approved.Int64
	row.RejectedRequests = rejected.Int64
	return row, nil
}

func (r *DashboardRepository) GetInvoiceSummary(ctx context.Context, projectIDs []uint64) (*InvoiceSummaryRow, error) {
	var query string
	var args []interface{}

	if len(projectIDs) > 0 {
		placeholders, pArgs := buildInClause(projectIDs)
		query = fmt.Sprintf(`SELECT COUNT(1), COALESCE(SUM(amount),0) FROM invoices WHERE project_id IN (%s)`, placeholders)
		args = pArgs
	} else {
		query = `SELECT COUNT(1), COALESCE(SUM(amount),0) FROM invoices`
	}

	row := &InvoiceSummaryRow{}
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&row.TotalInvoices, &row.TotalAmount)
	if err != nil {
		return nil, err
	}
	return row, nil
}
