package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/gilangrmdnii/invoice-backend/internal/model"
)

type ProjectPlanRepository struct {
	db *sql.DB
}

func NewProjectPlanRepository(db *sql.DB) *ProjectPlanRepository {
	return &ProjectPlanRepository{db: db}
}

func (r *ProjectPlanRepository) FindByProjectID(ctx context.Context, projectID uint64) ([]model.ProjectPlanItem, error) {
	query := `SELECT id, project_id, parent_id, is_label, description, quantity, unit, unit_price, subtotal, sort_order, created_at, updated_at
	FROM project_plan_items WHERE project_id = ? ORDER BY sort_order ASC, id ASC`

	rows, err := r.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.ProjectPlanItem
	for rows.Next() {
		var item model.ProjectPlanItem
		var parentID sql.NullInt64
		if err := rows.Scan(&item.ID, &item.ProjectID, &parentID, &item.IsLabel, &item.Description,
			&item.Quantity, &item.Unit, &item.UnitPrice, &item.Subtotal, &item.SortOrder,
			&item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		if parentID.Valid {
			v := uint64(parentID.Int64)
			item.ParentID = &v
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *ProjectPlanRepository) ReplaceAll(ctx context.Context, projectID uint64, items []model.ProjectPlanItem) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// Delete existing plan items
	_, err = tx.ExecContext(ctx, `DELETE FROM project_plan_items WHERE project_id = ?`, projectID)
	if err != nil {
		return fmt.Errorf("delete old plan items: %w", err)
	}

	var totalBudget float64

	// Insert new items (labels first with children, then standalone)
	for i, item := range items {
		if item.IsLabel {
			// Insert label row
			labelResult, err := tx.ExecContext(ctx,
				`INSERT INTO project_plan_items (project_id, parent_id, is_label, description, quantity, unit, unit_price, subtotal, sort_order)
				VALUES (?, NULL, TRUE, ?, 0, '', 0, 0, ?)`,
				projectID, item.Description, i,
			)
			if err != nil {
				return fmt.Errorf("insert plan label: %w", err)
			}
			labelID, err := labelResult.LastInsertId()
			if err != nil {
				return fmt.Errorf("get label id: %w", err)
			}
			// Insert children
			for j, child := range item.Children {
				subtotal := child.Quantity * child.UnitPrice
				totalBudget += subtotal
				_, err = tx.ExecContext(ctx,
					`INSERT INTO project_plan_items (project_id, parent_id, is_label, description, quantity, unit, unit_price, subtotal, sort_order)
					VALUES (?, ?, FALSE, ?, ?, ?, ?, ?, ?)`,
					projectID, labelID, child.Description, child.Quantity, child.Unit, child.UnitPrice, subtotal, j,
				)
				if err != nil {
					return fmt.Errorf("insert plan item under label: %w", err)
				}
			}
		} else {
			// Standalone item
			subtotal := item.Quantity * item.UnitPrice
			totalBudget += subtotal
			_, err = tx.ExecContext(ctx,
				`INSERT INTO project_plan_items (project_id, parent_id, is_label, description, quantity, unit, unit_price, subtotal, sort_order)
				VALUES (?, NULL, FALSE, ?, ?, ?, ?, ?, ?)`,
				projectID, item.Description, item.Quantity, item.Unit, item.UnitPrice, subtotal, i,
			)
			if err != nil {
				return fmt.Errorf("insert plan item: %w", err)
			}
		}
	}

	// Update project budget
	_, err = tx.ExecContext(ctx,
		`UPDATE project_budgets SET total_budget = ?, updated_at = NOW() WHERE project_id = ?`,
		totalBudget, projectID,
	)
	if err != nil {
		return fmt.Errorf("update budget: %w", err)
	}

	return tx.Commit()
}
