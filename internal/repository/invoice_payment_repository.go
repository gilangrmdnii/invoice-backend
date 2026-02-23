package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/gilangrmdnii/invoice-backend/internal/model"
)

type InvoicePaymentRepository struct {
	db *sql.DB
}

func NewInvoicePaymentRepository(db *sql.DB) *InvoicePaymentRepository {
	return &InvoicePaymentRepository{db: db}
}

func (r *InvoicePaymentRepository) Create(ctx context.Context, p *model.InvoicePayment) (uint64, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx,
		`INSERT INTO invoice_payments (invoice_id, amount, payment_date, payment_method, proof_url, notes, created_by)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		p.InvoiceID, p.Amount, p.PaymentDate, p.PaymentMethod, p.ProofURL, p.Notes, p.CreatedBy,
	)
	if err != nil {
		return 0, fmt.Errorf("insert payment: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	// Update invoice paid_amount and payment_status
	_, err = tx.ExecContext(ctx,
		`UPDATE invoices SET
			paid_amount = (SELECT COALESCE(SUM(amount), 0) FROM invoice_payments WHERE invoice_id = ?),
			payment_status = CASE
				WHEN (SELECT COALESCE(SUM(amount), 0) FROM invoice_payments WHERE invoice_id = ?) >= amount THEN 'PAID'
				WHEN (SELECT COALESCE(SUM(amount), 0) FROM invoice_payments WHERE invoice_id = ?) > 0 THEN 'PARTIAL_PAID'
				ELSE 'UNPAID'
			END
		WHERE id = ?`,
		p.InvoiceID, p.InvoiceID, p.InvoiceID, p.InvoiceID,
	)
	if err != nil {
		return 0, fmt.Errorf("update invoice payment status: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit tx: %w", err)
	}

	return uint64(id), nil
}

func (r *InvoicePaymentRepository) FindByInvoiceID(ctx context.Context, invoiceID uint64) ([]model.InvoicePayment, error) {
	query := `SELECT ip.id, ip.invoice_id, ip.amount, ip.payment_date, ip.payment_method,
		ip.proof_url, ip.notes, ip.created_by, ip.created_at, ip.updated_at
	FROM invoice_payments ip
	WHERE ip.invoice_id = ?
	ORDER BY ip.payment_date ASC, ip.created_at ASC`

	rows, err := r.db.QueryContext(ctx, query, invoiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []model.InvoicePayment
	for rows.Next() {
		var p model.InvoicePayment
		var proofURL, notes sql.NullString

		if err := rows.Scan(
			&p.ID, &p.InvoiceID, &p.Amount, &p.PaymentDate, &p.PaymentMethod,
			&proofURL, &notes, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}

		p.ProofURL = proofURL.String
		p.Notes = notes.String
		payments = append(payments, p)
	}
	return payments, rows.Err()
}

func (r *InvoicePaymentRepository) Delete(ctx context.Context, id, invoiceID uint64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `DELETE FROM invoice_payments WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete payment: %w", err)
	}

	// Recalculate invoice paid_amount and payment_status
	_, err = tx.ExecContext(ctx,
		`UPDATE invoices SET
			paid_amount = (SELECT COALESCE(SUM(amount), 0) FROM invoice_payments WHERE invoice_id = ?),
			payment_status = CASE
				WHEN (SELECT COALESCE(SUM(amount), 0) FROM invoice_payments WHERE invoice_id = ?) >= amount THEN 'PAID'
				WHEN (SELECT COALESCE(SUM(amount), 0) FROM invoice_payments WHERE invoice_id = ?) > 0 THEN 'PARTIAL_PAID'
				ELSE 'UNPAID'
			END
		WHERE id = ?`,
		invoiceID, invoiceID, invoiceID, invoiceID,
	)
	if err != nil {
		return fmt.Errorf("update invoice payment status: %w", err)
	}

	return tx.Commit()
}
