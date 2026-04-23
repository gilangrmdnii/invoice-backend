package repository

import (
	"context"
	"database/sql"

	"github.com/gilangrmdnii/invoice-backend/internal/model"
)

type FinanceReportRepository struct {
	db *sql.DB
}

func NewFinanceReportRepository(db *sql.DB) *FinanceReportRepository {
	return &FinanceReportRepository{db: db}
}

// DB exposes the underlying *sql.DB for custom queries (e.g., date groupings).
func (r *FinanceReportRepository) DB() *sql.DB {
	return r.db
}

// ============ Recruiter Fees ============

func (r *FinanceReportRepository) FindRecruiterFeesByProject(ctx context.Context, projectID uint64) ([]model.FinanceRecruiterFee, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, project_id, recruiter_name, jumlah, fee_recruiter,
			insentif_responden_main, jumlah_responden_main,
			insentif_responden_backup, jumlah_responden_backup,
			sort_order, created_at, updated_at
		FROM finance_recruiter_fees WHERE project_id = ?
		ORDER BY sort_order ASC, id ASC`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.FinanceRecruiterFee
	for rows.Next() {
		var f model.FinanceRecruiterFee
		if err := rows.Scan(&f.ID, &f.ProjectID, &f.RecruiterName, &f.Jumlah, &f.FeeRecruiter,
			&f.InsentifRespondenMain, &f.JumlahRespondenMain,
			&f.InsentifRespondenBackup, &f.JumlahRespondenBackup,
			&f.SortOrder, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

// ReplaceRecruiterFees: delete all then insert (transactional replace)
func (r *FinanceReportRepository) ReplaceRecruiterFees(ctx context.Context, projectID uint64, fees []model.FinanceRecruiterFee) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `DELETE FROM finance_recruiter_fees WHERE project_id = ?`, projectID); err != nil {
		return err
	}
	for _, f := range fees {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO finance_recruiter_fees (project_id, recruiter_name, jumlah, fee_recruiter,
				insentif_responden_main, jumlah_responden_main,
				insentif_responden_backup, jumlah_responden_backup, sort_order)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			projectID, f.RecruiterName, f.Jumlah, f.FeeRecruiter,
			f.InsentifRespondenMain, f.JumlahRespondenMain,
			f.InsentifRespondenBackup, f.JumlahRespondenBackup, f.SortOrder,
		); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// ============ Sample Entries ============

func (r *FinanceReportRepository) FindSampleEntriesByProject(ctx context.Context, projectID uint64) ([]model.FinanceSampleEntry, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, project_id, tanggal_pelaksanaan, jumlah_sample,
			insentif_responden_main, jumlah_responden_main,
			insentif_responden_backup, jumlah_responden_backup,
			sort_order, created_at, updated_at
		FROM finance_sample_entries WHERE project_id = ?
		ORDER BY tanggal_pelaksanaan ASC, sort_order ASC, id ASC`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.FinanceSampleEntry
	for rows.Next() {
		var s model.FinanceSampleEntry
		if err := rows.Scan(&s.ID, &s.ProjectID, &s.TanggalPelaksanaan, &s.JumlahSample,
			&s.InsentifRespondenMain, &s.JumlahRespondenMain,
			&s.InsentifRespondenBackup, &s.JumlahRespondenBackup,
			&s.SortOrder, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *FinanceReportRepository) ReplaceSampleEntries(ctx context.Context, projectID uint64, entries []model.FinanceSampleEntry) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `DELETE FROM finance_sample_entries WHERE project_id = ?`, projectID); err != nil {
		return err
	}
	for _, s := range entries {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO finance_sample_entries (project_id, tanggal_pelaksanaan, jumlah_sample,
				insentif_responden_main, jumlah_responden_main,
				insentif_responden_backup, jumlah_responden_backup, sort_order)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			projectID, s.TanggalPelaksanaan, s.JumlahSample,
			s.InsentifRespondenMain, s.JumlahRespondenMain,
			s.InsentifRespondenBackup, s.JumlahRespondenBackup, s.SortOrder,
		); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// ============ Manual Expenses ============

func (r *FinanceReportRepository) FindManualExpensesByProject(ctx context.Context, projectID uint64) ([]model.FinanceManualExpense, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, project_id, member_user_id, member_name, category, tanggal,
			description, quantity, unit_price, amount, sort_order,
			created_by, created_at, updated_at
		FROM finance_manual_expenses WHERE project_id = ?
		ORDER BY sort_order ASC, id ASC`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.FinanceManualExpense
	for rows.Next() {
		var e model.FinanceManualExpense
		if err := rows.Scan(&e.ID, &e.ProjectID, &e.MemberUserID, &e.MemberName, &e.Category, &e.Tanggal,
			&e.Description, &e.Quantity, &e.UnitPrice, &e.Amount, &e.SortOrder,
			&e.CreatedBy, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (r *FinanceReportRepository) ReplaceManualExpenses(ctx context.Context, projectID uint64, userID uint64, expenses []model.FinanceManualExpense) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `DELETE FROM finance_manual_expenses WHERE project_id = ?`, projectID); err != nil {
		return err
	}
	for _, e := range expenses {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO finance_manual_expenses (project_id, member_user_id, member_name, category, tanggal,
				description, quantity, unit_price, amount, sort_order, created_by)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			projectID, e.MemberUserID, e.MemberName, e.Category, e.Tanggal,
			e.Description, e.Quantity, e.UnitPrice, e.Amount, e.SortOrder, userID,
		); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// ============ Aggregations ============

// ExpensesByMemberCategory aggregates expenses table by creator (member) + category.
// Returns map[userID]map[category]amount.
func (r *FinanceReportRepository) AggregateExpenses(ctx context.Context, projectID uint64) ([]AggregatedExpense, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT e.created_by, u.full_name, u.role, e.category, COALESCE(SUM(e.amount), 0)
		FROM expenses e
		LEFT JOIN users u ON u.id = e.created_by
		WHERE e.project_id = ?
		GROUP BY e.created_by, u.full_name, u.role, e.category`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []AggregatedExpense
	for rows.Next() {
		var a AggregatedExpense
		if err := rows.Scan(&a.UserID, &a.FullName, &a.Role, &a.Category, &a.Amount); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

type AggregatedExpense struct {
	UserID   uint64
	FullName sql.NullString
	Role     sql.NullString
	Category string
	Amount   float64
}
