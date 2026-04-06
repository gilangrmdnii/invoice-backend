package repository

import (
	"context"
	"database/sql"

	"github.com/gilangrmdnii/invoice-backend/internal/model"
)

type ProjectWorkerRepository struct {
	db *sql.DB
}

func NewProjectWorkerRepository(db *sql.DB) *ProjectWorkerRepository {
	return &ProjectWorkerRepository{db: db}
}

func (r *ProjectWorkerRepository) Create(ctx context.Context, w *model.ProjectWorker) (uint64, error) {
	query := `INSERT INTO project_workers (project_id, full_name, role, phone, daily_wage, added_by) VALUES (?, ?, ?, ?, ?, ?)`
	result, err := r.db.ExecContext(ctx, query, w.ProjectID, w.FullName, w.Role, w.Phone, w.DailyWage, w.AddedBy)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return uint64(id), nil
}

func (r *ProjectWorkerRepository) FindByID(ctx context.Context, id uint64) (*model.ProjectWorker, error) {
	query := `SELECT id, project_id, full_name, role, phone, daily_wage, is_active, added_by, created_at, updated_at FROM project_workers WHERE id = ?`
	w := &model.ProjectWorker{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&w.ID, &w.ProjectID, &w.FullName, &w.Role, &w.Phone, &w.DailyWage, &w.IsActive, &w.AddedBy, &w.CreatedAt, &w.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return w, nil
}

func (r *ProjectWorkerRepository) FindByProjectID(ctx context.Context, projectID uint64) ([]model.ProjectWorker, error) {
	query := `SELECT id, project_id, full_name, role, phone, daily_wage, is_active, added_by, created_at, updated_at FROM project_workers WHERE project_id = ? ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var workers []model.ProjectWorker
	for rows.Next() {
		var w model.ProjectWorker
		if err := rows.Scan(&w.ID, &w.ProjectID, &w.FullName, &w.Role, &w.Phone, &w.DailyWage, &w.IsActive, &w.AddedBy, &w.CreatedAt, &w.UpdatedAt); err != nil {
			return nil, err
		}
		workers = append(workers, w)
	}
	return workers, rows.Err()
}

func (r *ProjectWorkerRepository) Update(ctx context.Context, w *model.ProjectWorker) error {
	query := `UPDATE project_workers SET full_name = ?, role = ?, phone = ?, daily_wage = ?, is_active = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, w.FullName, w.Role, w.Phone, w.DailyWage, w.IsActive, w.ID)
	return err
}

func (r *ProjectWorkerRepository) Delete(ctx context.Context, id uint64) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM project_workers WHERE id = ?`, id)
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
