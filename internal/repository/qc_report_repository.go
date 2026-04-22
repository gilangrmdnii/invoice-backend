package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/gilangrmdnii/invoice-backend/internal/model"
)

type QCReportRepository struct {
	db *sql.DB
}

func NewQCReportRepository(db *sql.DB) *QCReportRepository {
	return &QCReportRepository{db: db}
}

// ============ QC Report ============

func (r *QCReportRepository) Create(ctx context.Context, rep *model.QCReport, items []model.QCReportItem, recruiters []model.QCRecruiterPerformance) (uint64, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()

	result, err := tx.ExecContext(ctx, `
		INSERT INTO qc_reports (
			project_id, qc_user_id, spv_names, project_type, methodology, city, area,
			execution_start_date, execution_end_date, briefing_date, work_start_date, work_end_date,
			visit_target, visit_ok, telp_target, telp_ok, total_amount,
			status,
			location, report_date,
			qc_signatory_name, qc_signatory_title, coordinator_signatory_name, coordinator_signatory_title,
			note, created_by
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		rep.ProjectID, rep.QCUserID, rep.SPVNames, rep.ProjectType, rep.Methodology, rep.City, rep.Area,
		rep.ExecutionStartDate, rep.ExecutionEndDate, rep.BriefingDate, rep.WorkStartDate, rep.WorkEndDate,
		rep.VisitTarget, rep.VisitOK, rep.TelpTarget, rep.TelpOK, rep.TotalAmount,
		rep.Status,
		rep.Location, rep.ReportDate,
		rep.QCSignatoryName, rep.QCSignatoryTitle, rep.CoordinatorSignatoryName, rep.CoordinatorSignatoryTitle,
		rep.Note, rep.CreatedBy,
	)
	if err != nil {
		return 0, err
	}
	id64, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	id := uint64(id64)

	for _, it := range items {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO qc_report_items (qc_report_id, category, status, label, quantity, unit_price, subtotal, sort_order)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			id, it.Category, it.Status, it.Label, it.Quantity, it.UnitPrice, it.Subtotal, it.SortOrder,
		); err != nil {
			return 0, err
		}
	}

	for _, rc := range recruiters {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO qc_recruiter_performance (qc_report_id, recruiter_name, total, ok_perpi, do_perpi, ok_qc, do_qc, notes, sort_order)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			id, rc.RecruiterName, rc.Total, rc.OKPerpi, rc.DOPerpi, rc.OKQC, rc.DOQC, rc.Notes, rc.SortOrder,
		); err != nil {
			return 0, err
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return id, nil
}

func (r *QCReportRepository) FindByID(ctx context.Context, id uint64) (*model.QCReport, error) {
	rep := &model.QCReport{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, project_id, qc_user_id, spv_names, project_type, methodology, city, area,
			execution_start_date, execution_end_date, briefing_date, work_start_date, work_end_date,
			visit_target, visit_ok, telp_target, telp_ok, total_amount,
			status, approved_by, approval_notes, approved_at,
			location, report_date,
			qc_signatory_name, qc_signatory_title, coordinator_signatory_name, coordinator_signatory_title,
			note, created_by, created_at, updated_at
		FROM qc_reports WHERE id = ?`, id).Scan(
		&rep.ID, &rep.ProjectID, &rep.QCUserID, &rep.SPVNames, &rep.ProjectType, &rep.Methodology, &rep.City, &rep.Area,
		&rep.ExecutionStartDate, &rep.ExecutionEndDate, &rep.BriefingDate, &rep.WorkStartDate, &rep.WorkEndDate,
		&rep.VisitTarget, &rep.VisitOK, &rep.TelpTarget, &rep.TelpOK, &rep.TotalAmount,
		&rep.Status, &rep.ApprovedBy, &rep.ApprovalNotes, &rep.ApprovedAt,
		&rep.Location, &rep.ReportDate,
		&rep.QCSignatoryName, &rep.QCSignatoryTitle, &rep.CoordinatorSignatoryName, &rep.CoordinatorSignatoryTitle,
		&rep.Note, &rep.CreatedBy, &rep.CreatedAt, &rep.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return rep, nil
}

func (r *QCReportRepository) FindAll(ctx context.Context) ([]model.QCReport, error) {
	return r.queryList(ctx, `
		SELECT id, project_id, qc_user_id, spv_names, project_type, methodology, city, area,
			execution_start_date, execution_end_date, briefing_date, work_start_date, work_end_date,
			visit_target, visit_ok, telp_target, telp_ok, total_amount,
			status, approved_by, approval_notes, approved_at,
			location, report_date,
			qc_signatory_name, qc_signatory_title, coordinator_signatory_name, coordinator_signatory_title,
			note, created_by, created_at, updated_at
		FROM qc_reports ORDER BY created_at DESC`)
}

func (r *QCReportRepository) FindByProjectID(ctx context.Context, projectID uint64) ([]model.QCReport, error) {
	return r.queryList(ctx, `
		SELECT id, project_id, qc_user_id, spv_names, project_type, methodology, city, area,
			execution_start_date, execution_end_date, briefing_date, work_start_date, work_end_date,
			visit_target, visit_ok, telp_target, telp_ok, total_amount,
			status, approved_by, approval_notes, approved_at,
			location, report_date,
			qc_signatory_name, qc_signatory_title, coordinator_signatory_name, coordinator_signatory_title,
			note, created_by, created_at, updated_at
		FROM qc_reports WHERE project_id = ? ORDER BY created_at DESC`, projectID)
}

func (r *QCReportRepository) FindByProjectIDs(ctx context.Context, projectIDs []uint64) ([]model.QCReport, error) {
	if len(projectIDs) == 0 {
		return nil, nil
	}
	placeholders, args := buildInClause(projectIDs)
	query := fmt.Sprintf(`
		SELECT id, project_id, qc_user_id, spv_names, project_type, methodology, city, area,
			execution_start_date, execution_end_date, briefing_date, work_start_date, work_end_date,
			visit_target, visit_ok, telp_target, telp_ok, total_amount,
			status, approved_by, approval_notes, approved_at,
			location, report_date,
			qc_signatory_name, qc_signatory_title, coordinator_signatory_name, coordinator_signatory_title,
			note, created_by, created_at, updated_at
		FROM qc_reports WHERE project_id IN (%s) ORDER BY created_at DESC`, placeholders)
	return r.queryList(ctx, query, args...)
}

func (r *QCReportRepository) queryList(ctx context.Context, query string, args ...interface{}) ([]model.QCReport, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.QCReport
	for rows.Next() {
		var rep model.QCReport
		if err := rows.Scan(
			&rep.ID, &rep.ProjectID, &rep.QCUserID, &rep.SPVNames, &rep.ProjectType, &rep.Methodology, &rep.City, &rep.Area,
			&rep.ExecutionStartDate, &rep.ExecutionEndDate, &rep.BriefingDate, &rep.WorkStartDate, &rep.WorkEndDate,
			&rep.VisitTarget, &rep.VisitOK, &rep.TelpTarget, &rep.TelpOK, &rep.TotalAmount,
			&rep.Location, &rep.ReportDate,
			&rep.QCSignatoryName, &rep.QCSignatoryTitle, &rep.CoordinatorSignatoryName, &rep.CoordinatorSignatoryTitle,
			&rep.Note, &rep.CreatedBy, &rep.CreatedAt, &rep.UpdatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, rep)
	}
	return out, rows.Err()
}

func (r *QCReportRepository) Update(ctx context.Context, rep *model.QCReport, items []model.QCReportItem, recruiters []model.QCRecruiterPerformance) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `
		UPDATE qc_reports SET
			qc_user_id = ?, spv_names = ?, project_type = ?, methodology = ?, city = ?, area = ?,
			execution_start_date = ?, execution_end_date = ?, briefing_date = ?, work_start_date = ?, work_end_date = ?,
			visit_target = ?, visit_ok = ?, telp_target = ?, telp_ok = ?, total_amount = ?,
			status = ?,
			location = ?, report_date = ?,
			qc_signatory_name = ?, qc_signatory_title = ?, coordinator_signatory_name = ?, coordinator_signatory_title = ?,
			note = ?
		WHERE id = ?`,
		rep.QCUserID, rep.SPVNames, rep.ProjectType, rep.Methodology, rep.City, rep.Area,
		rep.ExecutionStartDate, rep.ExecutionEndDate, rep.BriefingDate, rep.WorkStartDate, rep.WorkEndDate,
		rep.VisitTarget, rep.VisitOK, rep.TelpTarget, rep.TelpOK, rep.TotalAmount,
		rep.Status,
		rep.Location, rep.ReportDate,
		rep.QCSignatoryName, rep.QCSignatoryTitle, rep.CoordinatorSignatoryName, rep.CoordinatorSignatoryTitle,
		rep.Note, rep.ID,
	); err != nil {
		return err
	}

	// Replace items
	if _, err := tx.ExecContext(ctx, `DELETE FROM qc_report_items WHERE qc_report_id = ?`, rep.ID); err != nil {
		return err
	}
	for _, it := range items {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO qc_report_items (qc_report_id, category, status, label, quantity, unit_price, subtotal, sort_order)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			rep.ID, it.Category, it.Status, it.Label, it.Quantity, it.UnitPrice, it.Subtotal, it.SortOrder,
		); err != nil {
			return err
		}
	}

	// Replace recruiters
	if _, err := tx.ExecContext(ctx, `DELETE FROM qc_recruiter_performance WHERE qc_report_id = ?`, rep.ID); err != nil {
		return err
	}
	for _, rc := range recruiters {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO qc_recruiter_performance (qc_report_id, recruiter_name, total, ok_perpi, do_perpi, ok_qc, do_qc, notes, sort_order)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			rep.ID, rc.RecruiterName, rc.Total, rc.OKPerpi, rc.DOPerpi, rc.OKQC, rc.DOQC, rc.Notes, rc.SortOrder,
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *QCReportRepository) Delete(ctx context.Context, id uint64) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM qc_reports WHERE id = ?`, id)
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

// SetApproval updates approval status, approver, notes, and approved_at
func (r *QCReportRepository) SetApproval(ctx context.Context, id uint64, status model.QCReportStatus, approvedBy uint64, notes string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE qc_reports SET status = ?, approved_by = ?, approval_notes = ?, approved_at = CURRENT_TIMESTAMP
		WHERE id = ?`,
		status, approvedBy, notes, id,
	)
	return err
}

// ============ Items & Recruiters ============

func (r *QCReportRepository) FindItems(ctx context.Context, reportID uint64) ([]model.QCReportItem, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, qc_report_id, category, status, label, quantity, unit_price, subtotal, sort_order, created_at
		FROM qc_report_items WHERE qc_report_id = ? ORDER BY sort_order ASC, id ASC`, reportID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.QCReportItem
	for rows.Next() {
		var it model.QCReportItem
		if err := rows.Scan(&it.ID, &it.QCReportID, &it.Category, &it.Status, &it.Label, &it.Quantity, &it.UnitPrice, &it.Subtotal, &it.SortOrder, &it.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

func (r *QCReportRepository) FindRecruiters(ctx context.Context, reportID uint64) ([]model.QCRecruiterPerformance, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, qc_report_id, recruiter_name, total, ok_perpi, do_perpi, ok_qc, do_qc, notes, sort_order, created_at
		FROM qc_recruiter_performance WHERE qc_report_id = ? ORDER BY sort_order ASC, id ASC`, reportID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.QCRecruiterPerformance
	for rows.Next() {
		var rc model.QCRecruiterPerformance
		if err := rows.Scan(&rc.ID, &rc.QCReportID, &rc.RecruiterName, &rc.Total, &rc.OKPerpi, &rc.DOPerpi, &rc.OKQC, &rc.DOQC, &rc.Notes, &rc.SortOrder, &rc.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, rc)
	}
	return out, rows.Err()
}
