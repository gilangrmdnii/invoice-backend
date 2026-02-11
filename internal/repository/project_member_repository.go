package repository

import (
	"context"
	"database/sql"

	"github.com/gilangrmdnii/invoice-backend/internal/model"
)

type ProjectMemberRepository struct {
	db *sql.DB
}

func NewProjectMemberRepository(db *sql.DB) *ProjectMemberRepository {
	return &ProjectMemberRepository{db: db}
}

func (r *ProjectMemberRepository) Create(ctx context.Context, member *model.ProjectMember) (uint64, error) {
	query := `INSERT INTO project_members (project_id, user_id) VALUES (?, ?)`
	result, err := r.db.ExecContext(ctx, query, member.ProjectID, member.UserID)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return uint64(id), nil
}

func (r *ProjectMemberRepository) Delete(ctx context.Context, projectID, userID uint64) error {
	query := `DELETE FROM project_members WHERE project_id = ? AND user_id = ?`
	result, err := r.db.ExecContext(ctx, query, projectID, userID)
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

func (r *ProjectMemberRepository) FindByProjectID(ctx context.Context, projectID uint64) ([]model.ProjectMember, error) {
	query := `SELECT id, project_id, user_id, created_at FROM project_members WHERE project_id = ?`
	rows, err := r.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []model.ProjectMember
	for rows.Next() {
		var m model.ProjectMember
		if err := rows.Scan(&m.ID, &m.ProjectID, &m.UserID, &m.CreatedAt); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, rows.Err()
}

func (r *ProjectMemberRepository) Exists(ctx context.Context, projectID, userID uint64) (bool, error) {
	query := `SELECT COUNT(1) FROM project_members WHERE project_id = ? AND user_id = ?`
	var count int
	err := r.db.QueryRowContext(ctx, query, projectID, userID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
