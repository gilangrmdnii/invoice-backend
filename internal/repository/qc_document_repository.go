package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/gilangrmdnii/invoice-backend/internal/model"
)

type QCDocumentRepository struct {
	db *sql.DB
}

func NewQCDocumentRepository(db *sql.DB) *QCDocumentRepository {
	return &QCDocumentRepository{db: db}
}

func (r *QCDocumentRepository) Create(ctx context.Context, doc *model.QCDocument) (uint64, error) {
	query := `INSERT INTO qc_documents (project_id, title, description, document_type, file_url, uploaded_by) VALUES (?, ?, ?, ?, ?, ?)`
	result, err := r.db.ExecContext(ctx, query, doc.ProjectID, doc.Title, doc.Description, doc.DocumentType, doc.FileURL, doc.UploadedBy)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return uint64(id), nil
}

func (r *QCDocumentRepository) FindByID(ctx context.Context, id uint64) (*model.QCDocument, error) {
	query := `SELECT id, project_id, title, description, document_type, file_url, uploaded_by, created_at, updated_at FROM qc_documents WHERE id = ?`
	doc := &model.QCDocument{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&doc.ID, &doc.ProjectID, &doc.Title, &doc.Description, &doc.DocumentType, &doc.FileURL, &doc.UploadedBy, &doc.CreatedAt, &doc.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return doc, nil
}

func (r *QCDocumentRepository) FindByProjectID(ctx context.Context, projectID uint64) ([]model.QCDocument, error) {
	query := `SELECT id, project_id, title, description, document_type, file_url, uploaded_by, created_at, updated_at FROM qc_documents WHERE project_id = ? ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var docs []model.QCDocument
	for rows.Next() {
		var d model.QCDocument
		if err := rows.Scan(&d.ID, &d.ProjectID, &d.Title, &d.Description, &d.DocumentType, &d.FileURL, &d.UploadedBy, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		docs = append(docs, d)
	}
	return docs, rows.Err()
}

func (r *QCDocumentRepository) FindByProjectIDs(ctx context.Context, projectIDs []uint64) ([]model.QCDocument, error) {
	if len(projectIDs) == 0 {
		return nil, nil
	}
	placeholders, args := buildInClause(projectIDs)
	query := fmt.Sprintf(`SELECT id, project_id, title, description, document_type, file_url, uploaded_by, created_at, updated_at FROM qc_documents WHERE project_id IN (%s) ORDER BY created_at DESC`, placeholders)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var docs []model.QCDocument
	for rows.Next() {
		var d model.QCDocument
		if err := rows.Scan(&d.ID, &d.ProjectID, &d.Title, &d.Description, &d.DocumentType, &d.FileURL, &d.UploadedBy, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		docs = append(docs, d)
	}
	return docs, rows.Err()
}

func (r *QCDocumentRepository) FindAll(ctx context.Context) ([]model.QCDocument, error) {
	query := `SELECT id, project_id, title, description, document_type, file_url, uploaded_by, created_at, updated_at FROM qc_documents ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var docs []model.QCDocument
	for rows.Next() {
		var d model.QCDocument
		if err := rows.Scan(&d.ID, &d.ProjectID, &d.Title, &d.Description, &d.DocumentType, &d.FileURL, &d.UploadedBy, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		docs = append(docs, d)
	}
	return docs, rows.Err()
}

func (r *QCDocumentRepository) Update(ctx context.Context, doc *model.QCDocument) error {
	query := `UPDATE qc_documents SET title = ?, description = ?, document_type = ?, file_url = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, doc.Title, doc.Description, doc.DocumentType, doc.FileURL, doc.ID)
	return err
}

func (r *QCDocumentRepository) Delete(ctx context.Context, id uint64) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM qc_documents WHERE id = ?`, id)
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
