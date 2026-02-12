package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/gilangrmdnii/invoice-backend/internal/model"
)

type InvoiceRepository struct {
	db *sql.DB
}

func NewInvoiceRepository(db *sql.DB) *InvoiceRepository {
	return &InvoiceRepository{db: db}
}

func (r *InvoiceRepository) Create(ctx context.Context, inv *model.Invoice) (uint64, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	tempNumber := fmt.Sprintf("TEMP-%d", time.Now().UnixNano())
	result, err := tx.ExecContext(ctx,
		`INSERT INTO invoices (invoice_number, project_id, amount, file_url, created_by) VALUES (?, ?, ?, ?, ?)`,
		tempNumber, inv.ProjectID, inv.Amount, inv.FileURL, inv.CreatedBy,
	)
	if err != nil {
		return 0, fmt.Errorf("insert invoice: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	invoiceNumber := fmt.Sprintf("INV-%s-%04d", time.Now().Format("20060102"), id)
	_, err = tx.ExecContext(ctx, `UPDATE invoices SET invoice_number = ? WHERE id = ?`, invoiceNumber, id)
	if err != nil {
		return 0, fmt.Errorf("update invoice number: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit tx: %w", err)
	}

	inv.InvoiceNumber = invoiceNumber
	return uint64(id), nil
}

func (r *InvoiceRepository) FindByID(ctx context.Context, id uint64) (*model.Invoice, error) {
	query := `SELECT id, invoice_number, project_id, amount, file_url, created_by, created_at, updated_at FROM invoices WHERE id = ?`
	inv := &model.Invoice{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&inv.ID, &inv.InvoiceNumber, &inv.ProjectID, &inv.Amount, &inv.FileURL, &inv.CreatedBy, &inv.CreatedAt, &inv.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return inv, nil
}

func (r *InvoiceRepository) FindAll(ctx context.Context) ([]model.Invoice, error) {
	query := `SELECT id, invoice_number, project_id, amount, file_url, created_by, created_at, updated_at FROM invoices ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanInvoices(rows)
}

func (r *InvoiceRepository) FindByProjectIDs(ctx context.Context, projectIDs []uint64) ([]model.Invoice, error) {
	if len(projectIDs) == 0 {
		return nil, nil
	}
	placeholders, args := buildInClause(projectIDs)
	query := fmt.Sprintf(`SELECT id, invoice_number, project_id, amount, file_url, created_by, created_at, updated_at FROM invoices WHERE project_id IN (%s) ORDER BY created_at DESC`, placeholders)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanInvoices(rows)
}

func (r *InvoiceRepository) Update(ctx context.Context, inv *model.Invoice) error {
	query := `UPDATE invoices SET amount = ?, file_url = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, inv.Amount, inv.FileURL, inv.ID)
	return err
}

func (r *InvoiceRepository) Delete(ctx context.Context, id uint64) error {
	query := `DELETE FROM invoices WHERE id = ?`
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

func scanInvoices(rows *sql.Rows) ([]model.Invoice, error) {
	var invoices []model.Invoice
	for rows.Next() {
		var inv model.Invoice
		if err := rows.Scan(&inv.ID, &inv.InvoiceNumber, &inv.ProjectID, &inv.Amount, &inv.FileURL, &inv.CreatedBy, &inv.CreatedAt, &inv.UpdatedAt); err != nil {
			return nil, err
		}
		invoices = append(invoices, inv)
	}
	return invoices, rows.Err()
}
