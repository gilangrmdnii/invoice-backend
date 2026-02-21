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

func (r *InvoiceRepository) Create(ctx context.Context, inv *model.Invoice, items []model.InvoiceItem) (uint64, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// Insert invoice with temp number
	tempNumber := fmt.Sprintf("TEMP-%d", time.Now().UnixNano())
	result, err := tx.ExecContext(ctx,
		`INSERT INTO invoices (invoice_number, invoice_type, project_id, amount, status, file_url,
			recipient_name, recipient_address, attention, po_number, invoice_date,
			dp_percentage, subtotal, tax_percentage, tax_amount, notes, language, created_by)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		tempNumber, inv.InvoiceType, inv.ProjectID, inv.Amount, model.InvoiceStatusPending, inv.FileURL,
		inv.RecipientName, inv.RecipientAddress, inv.Attention, inv.PONumber, inv.InvoiceDate,
		inv.DPPercentage, inv.Subtotal, inv.TaxPercentage, inv.TaxAmount, inv.Notes, inv.Language, inv.CreatedBy,
	)
	if err != nil {
		return 0, fmt.Errorf("insert invoice: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	// Generate invoice number: {seq}/INV/{company_code}/{MM}/{YYYY}
	var companyCode string
	err = tx.QueryRowContext(ctx, `SELECT company_code FROM company_settings LIMIT 1`).Scan(&companyCode)
	if err != nil {
		companyCode = "INV" // fallback
	}

	invoiceNumber := fmt.Sprintf("%03d/INV/%s/%s/%s",
		id,
		companyCode,
		inv.InvoiceDate.Format("01"),
		inv.InvoiceDate.Format("2006"),
	)
	_, err = tx.ExecContext(ctx, `UPDATE invoices SET invoice_number = ? WHERE id = ?`, invoiceNumber, id)
	if err != nil {
		return 0, fmt.Errorf("update invoice number: %w", err)
	}
	inv.InvoiceNumber = invoiceNumber

	// Insert invoice items (labels first, then children)
	for i, item := range items {
		if item.IsLabel {
			// Insert label row
			labelResult, err := tx.ExecContext(ctx,
				`INSERT INTO invoice_items (invoice_id, parent_id, is_label, description, quantity, unit, unit_price, subtotal, sort_order)
				VALUES (?, NULL, TRUE, ?, 0, '', 0, 0, ?)`,
				id, item.Description, i,
			)
			if err != nil {
				return 0, fmt.Errorf("insert invoice label: %w", err)
			}
			labelID, err := labelResult.LastInsertId()
			if err != nil {
				return 0, fmt.Errorf("get label id: %w", err)
			}
			// Insert children for this label
			for j, child := range item.Children {
				_, err = tx.ExecContext(ctx,
					`INSERT INTO invoice_items (invoice_id, parent_id, is_label, description, quantity, unit, unit_price, subtotal, sort_order)
					VALUES (?, ?, FALSE, ?, ?, ?, ?, ?, ?)`,
					id, labelID, child.Description, child.Quantity, child.Unit, child.UnitPrice, child.Subtotal, j,
				)
				if err != nil {
					return 0, fmt.Errorf("insert invoice item under label: %w", err)
				}
			}
		} else {
			// Standalone item (no label parent)
			_, err = tx.ExecContext(ctx,
				`INSERT INTO invoice_items (invoice_id, parent_id, is_label, description, quantity, unit, unit_price, subtotal, sort_order)
				VALUES (?, NULL, FALSE, ?, ?, ?, ?, ?, ?)`,
				id, item.Description, item.Quantity, item.Unit, item.UnitPrice, item.Subtotal, i,
			)
			if err != nil {
				return 0, fmt.Errorf("insert invoice item: %w", err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit tx: %w", err)
	}

	return uint64(id), nil
}

func (r *InvoiceRepository) FindByID(ctx context.Context, id uint64) (*model.Invoice, error) {
	query := `SELECT id, invoice_number, invoice_type, project_id, amount, status, file_url,
		recipient_name, recipient_address, attention, po_number, invoice_date,
		dp_percentage, subtotal, tax_percentage, tax_amount, notes, language,
		created_by, approved_by, reject_notes, created_at, updated_at
	FROM invoices WHERE id = ?`

	inv := &model.Invoice{}
	var fileURL, recipientAddr, attention, poNumber, notes, rejectNotes sql.NullString
	var dpPercentage sql.NullFloat64
	var approvedBy sql.NullInt64

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&inv.ID, &inv.InvoiceNumber, &inv.InvoiceType, &inv.ProjectID, &inv.Amount, &inv.Status, &fileURL,
		&inv.RecipientName, &recipientAddr, &attention, &poNumber, &inv.InvoiceDate,
		&dpPercentage, &inv.Subtotal, &inv.TaxPercentage, &inv.TaxAmount, &notes, &inv.Language,
		&inv.CreatedBy, &approvedBy, &rejectNotes, &inv.CreatedAt, &inv.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	inv.FileURL = fileURL.String
	inv.RecipientAddress = recipientAddr.String
	inv.Attention = attention.String
	inv.PONumber = poNumber.String
	inv.Notes = notes.String
	inv.RejectNotes = rejectNotes.String
	if dpPercentage.Valid {
		inv.DPPercentage = &dpPercentage.Float64
	}
	if approvedBy.Valid {
		v := uint64(approvedBy.Int64)
		inv.ApprovedBy = &v
	}

	return inv, nil
}

func (r *InvoiceRepository) FindAll(ctx context.Context) ([]model.Invoice, error) {
	query := `SELECT id, invoice_number, invoice_type, project_id, amount, status, file_url,
		recipient_name, recipient_address, attention, po_number, invoice_date,
		dp_percentage, subtotal, tax_percentage, tax_amount, notes, language,
		created_by, approved_by, reject_notes, created_at, updated_at
	FROM invoices ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanInvoices(rows)
}

func (r *InvoiceRepository) FindByProjectIDs(ctx context.Context, projectIDs []uint64) ([]model.Invoice, error) {
	if len(projectIDs) == 0 {
		return nil, nil
	}
	placeholders, args := buildInClause(projectIDs)
	query := fmt.Sprintf(`SELECT id, invoice_number, invoice_type, project_id, amount, status, file_url,
		recipient_name, recipient_address, attention, po_number, invoice_date,
		dp_percentage, subtotal, tax_percentage, tax_amount, notes, language,
		created_by, approved_by, reject_notes, created_at, updated_at
	FROM invoices WHERE project_id IN (%s) ORDER BY created_at DESC`, placeholders)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanInvoices(rows)
}

func (r *InvoiceRepository) FindItemsByInvoiceID(ctx context.Context, invoiceID uint64) ([]model.InvoiceItem, error) {
	query := `SELECT id, invoice_id, parent_id, is_label, description, quantity, unit, unit_price, subtotal, sort_order, created_at
	FROM invoice_items WHERE invoice_id = ? ORDER BY sort_order ASC`

	rows, err := r.db.QueryContext(ctx, query, invoiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.InvoiceItem
	for rows.Next() {
		var item model.InvoiceItem
		var parentID sql.NullInt64
		if err := rows.Scan(&item.ID, &item.InvoiceID, &parentID, &item.IsLabel, &item.Description, &item.Quantity,
			&item.Unit, &item.UnitPrice, &item.Subtotal, &item.SortOrder, &item.CreatedAt); err != nil {
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

func (r *InvoiceRepository) Update(ctx context.Context, inv *model.Invoice, items []model.InvoiceItem) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx,
		`UPDATE invoices SET recipient_name = ?, recipient_address = ?, attention = ?,
			po_number = ?, invoice_date = ?, dp_percentage = ?, subtotal = ?,
			tax_percentage = ?, tax_amount = ?, amount = ?, notes = ?, language = ?, file_url = ?
		WHERE id = ?`,
		inv.RecipientName, inv.RecipientAddress, inv.Attention,
		inv.PONumber, inv.InvoiceDate, inv.DPPercentage, inv.Subtotal,
		inv.TaxPercentage, inv.TaxAmount, inv.Amount, inv.Notes, inv.Language, inv.FileURL,
		inv.ID,
	)
	if err != nil {
		return fmt.Errorf("update invoice: %w", err)
	}

	// Replace items if provided
	if items != nil {
		_, err = tx.ExecContext(ctx, `DELETE FROM invoice_items WHERE invoice_id = ?`, inv.ID)
		if err != nil {
			return fmt.Errorf("delete old items: %w", err)
		}
		for i, item := range items {
			if item.IsLabel {
				labelResult, err := tx.ExecContext(ctx,
					`INSERT INTO invoice_items (invoice_id, parent_id, is_label, description, quantity, unit, unit_price, subtotal, sort_order)
					VALUES (?, NULL, TRUE, ?, 0, '', 0, 0, ?)`,
					inv.ID, item.Description, i,
				)
				if err != nil {
					return fmt.Errorf("insert invoice label: %w", err)
				}
				labelID, err := labelResult.LastInsertId()
				if err != nil {
					return fmt.Errorf("get label id: %w", err)
				}
				for j, child := range item.Children {
					_, err = tx.ExecContext(ctx,
						`INSERT INTO invoice_items (invoice_id, parent_id, is_label, description, quantity, unit, unit_price, subtotal, sort_order)
						VALUES (?, ?, FALSE, ?, ?, ?, ?, ?, ?)`,
						inv.ID, labelID, child.Description, child.Quantity, child.Unit, child.UnitPrice, child.Subtotal, j,
					)
					if err != nil {
						return fmt.Errorf("insert invoice item under label: %w", err)
					}
				}
			} else {
				_, err = tx.ExecContext(ctx,
					`INSERT INTO invoice_items (invoice_id, parent_id, is_label, description, quantity, unit, unit_price, subtotal, sort_order)
					VALUES (?, NULL, FALSE, ?, ?, ?, ?, ?, ?)`,
					inv.ID, item.Description, item.Quantity, item.Unit, item.UnitPrice, item.Subtotal, i,
				)
				if err != nil {
					return fmt.Errorf("insert invoice item: %w", err)
				}
			}
		}
	}

	return tx.Commit()
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

func (r *InvoiceRepository) ApproveInvoice(ctx context.Context, invoiceID, approvedBy uint64, notes string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	var status model.InvoiceStatus
	err = tx.QueryRowContext(ctx,
		`SELECT status FROM invoices WHERE id = ? FOR UPDATE`, invoiceID,
	).Scan(&status)
	if err != nil {
		return err
	}
	if status != model.InvoiceStatusPending {
		return fmt.Errorf("invoice is not pending")
	}

	_, err = tx.ExecContext(ctx,
		`UPDATE invoices SET status = 'APPROVED', approved_by = ? WHERE id = ?`,
		approvedBy, invoiceID,
	)
	if err != nil {
		return fmt.Errorf("update invoice: %w", err)
	}

	return tx.Commit()
}

func (r *InvoiceRepository) RejectInvoice(ctx context.Context, invoiceID, rejectedBy uint64, notes string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	var status model.InvoiceStatus
	err = tx.QueryRowContext(ctx,
		`SELECT status FROM invoices WHERE id = ? FOR UPDATE`, invoiceID,
	).Scan(&status)
	if err != nil {
		return err
	}
	if status != model.InvoiceStatusPending {
		return fmt.Errorf("invoice is not pending")
	}

	_, err = tx.ExecContext(ctx,
		`UPDATE invoices SET status = 'REJECTED', approved_by = ?, reject_notes = ? WHERE id = ?`,
		rejectedBy, notes, invoiceID,
	)
	if err != nil {
		return fmt.Errorf("update invoice: %w", err)
	}

	return tx.Commit()
}

func (r *InvoiceRepository) scanInvoices(rows *sql.Rows) ([]model.Invoice, error) {
	var invoices []model.Invoice
	for rows.Next() {
		var inv model.Invoice
		var fileURL, recipientAddr, attention, poNumber, notes, rejectNotes sql.NullString
		var dpPercentage sql.NullFloat64
		var approvedBy sql.NullInt64

		if err := rows.Scan(
			&inv.ID, &inv.InvoiceNumber, &inv.InvoiceType, &inv.ProjectID, &inv.Amount, &inv.Status, &fileURL,
			&inv.RecipientName, &recipientAddr, &attention, &poNumber, &inv.InvoiceDate,
			&dpPercentage, &inv.Subtotal, &inv.TaxPercentage, &inv.TaxAmount, &notes, &inv.Language,
			&inv.CreatedBy, &approvedBy, &rejectNotes, &inv.CreatedAt, &inv.UpdatedAt,
		); err != nil {
			return nil, err
		}

		inv.FileURL = fileURL.String
		inv.RecipientAddress = recipientAddr.String
		inv.Attention = attention.String
		inv.PONumber = poNumber.String
		inv.Notes = notes.String
		inv.RejectNotes = rejectNotes.String
		if dpPercentage.Valid {
			inv.DPPercentage = &dpPercentage.Float64
		}
		if approvedBy.Valid {
			v := uint64(approvedBy.Int64)
			inv.ApprovedBy = &v
		}

		invoices = append(invoices, inv)
	}
	return invoices, rows.Err()
}
