package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/gilangrmdnii/invoice-backend/internal/model"
)

type ProjectRepository struct {
	db *sql.DB
}

func NewProjectRepository(db *sql.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

func (r *ProjectRepository) CreateWithBudget(ctx context.Context, project *model.Project, totalBudget float64) (uint64, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx,
		`INSERT INTO projects (name, description, status, created_by) VALUES (?, ?, ?, ?)`,
		project.Name, project.Description, project.Status, project.CreatedBy,
	)
	if err != nil {
		return 0, fmt.Errorf("insert project: %w", err)
	}

	projectID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO project_budgets (project_id, total_budget) VALUES (?, ?)`,
		projectID, totalBudget,
	)
	if err != nil {
		return 0, fmt.Errorf("insert budget: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit tx: %w", err)
	}

	return uint64(projectID), nil
}

func (r *ProjectRepository) FindByID(ctx context.Context, id uint64) (*model.Project, error) {
	query := `SELECT id, name, description, status, created_by, created_at, updated_at FROM projects WHERE id = ?`
	p := &model.Project{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&p.ID, &p.Name, &p.Description, &p.Status, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *ProjectRepository) FindAll(ctx context.Context) ([]model.Project, error) {
	query := `SELECT id, name, description, status, created_by, created_at, updated_at FROM projects ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []model.Project
	for rows.Next() {
		var p model.Project
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Status, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, rows.Err()
}

func (r *ProjectRepository) FindByMemberUserID(ctx context.Context, userID uint64) ([]model.Project, error) {
	query := `SELECT p.id, p.name, p.description, p.status, p.created_by, p.created_at, p.updated_at
		FROM projects p
		INNER JOIN project_members pm ON p.id = pm.project_id
		WHERE pm.user_id = ?
		ORDER BY p.created_at DESC`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []model.Project
	for rows.Next() {
		var p model.Project
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Status, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, rows.Err()
}

func (r *ProjectRepository) Update(ctx context.Context, project *model.Project) error {
	query := `UPDATE projects SET name = ?, description = ?, status = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, project.Name, project.Description, project.Status, project.ID)
	return err
}
