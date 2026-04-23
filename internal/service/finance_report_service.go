package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gilangrmdnii/invoice-backend/internal/dto/request"
	"github.com/gilangrmdnii/invoice-backend/internal/dto/response"
	"github.com/gilangrmdnii/invoice-backend/internal/model"
	"github.com/gilangrmdnii/invoice-backend/internal/repository"
)

type FinanceReportService struct {
	reportRepo    *repository.FinanceReportRepository
	projectRepo   *repository.ProjectRepository
	memberRepo    *repository.ProjectMemberRepository
	expenseRepo   *repository.ExpenseRepository
	qcReportRepo  *repository.QCReportRepository
	userRepo      *repository.UserRepository
	auditRepo     *repository.AuditLogRepository
}

func NewFinanceReportService(
	reportRepo *repository.FinanceReportRepository,
	projectRepo *repository.ProjectRepository,
	memberRepo *repository.ProjectMemberRepository,
	expenseRepo *repository.ExpenseRepository,
	qcReportRepo *repository.QCReportRepository,
	userRepo *repository.UserRepository,
	auditRepo *repository.AuditLogRepository,
) *FinanceReportService {
	return &FinanceReportService{
		reportRepo:   reportRepo,
		projectRepo:  projectRepo,
		memberRepo:   memberRepo,
		expenseRepo:  expenseRepo,
		qcReportRepo: qcReportRepo,
		userRepo:     userRepo,
		auditRepo:    auditRepo,
	}
}

// Get builds the full Finance Report aggregation
func (s *FinanceReportService) Get(ctx context.Context, projectID uint64) (*response.FinanceReportResponse, error) {
	project, err := s.projectRepo.FindByID(ctx, projectID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("project not found")
		}
		return nil, err
	}

	resp := &response.FinanceReportResponse{
		ProjectID:        projectID,
		ProjectName:      project.Name,
		MemberBreakdowns: []response.MemberBreakdownResponse{},
		DailyExpenses:    []response.DateExpenseRow{},
		RecruiterFees:    []response.FinanceRecruiterFeeResponse{},
		SampleEntries:    []response.FinanceSampleEntryResponse{},
		ManualExpenses:   []response.FinanceManualExpenseResponse{},
	}

	// --- Project members (SPV + QC) ---
	members, err := s.memberRepo.FindByProjectID(ctx, projectID)
	if err != nil {
		return nil, err
	}
	type memberInfo struct {
		UserID   uint64
		FullName string
		Role     string
	}
	memberByID := make(map[uint64]memberInfo, len(members))
	spvNames, qcNames := []string{}, []string{}
	for _, m := range members {
		u, _ := s.userRepo.FindByID(ctx, m.UserID)
		info := memberInfo{UserID: m.UserID}
		if u != nil {
			info.FullName = u.FullName
			info.Role = string(u.Role)
		}
		memberByID[m.UserID] = info
		if info.Role == "SPV" {
			spvNames = append(spvNames, info.FullName)
		}
		if info.Role == "QC" || info.Role == "QC_COORDINATOR" {
			qcNames = append(qcNames, info.FullName)
		}
	}
	resp.SPVNames = strings.Join(spvNames, ", ")
	resp.QCNames = strings.Join(qcNames, ", ")

	// --- Execution date range (from latest QC report for this project, if any) ---
	qcReports, _ := s.qcReportRepo.FindByProjectID(ctx, projectID)
	if len(qcReports) > 0 {
		last := qcReports[0] // ordered by created_at DESC
		resp.ExecutionStartDate = last.ExecutionStartDate
		resp.ExecutionEndDate = last.ExecutionEndDate
	}

	// --- Aggregate expenses per member per category ---
	aggs, err := s.reportRepo.AggregateExpenses(ctx, projectID)
	if err != nil {
		return nil, err
	}
	memberBreakdown := make(map[uint64]*response.MemberBreakdownResponse)
	// Init breakdown for all project members (even if no expense yet)
	for _, m := range memberByID {
		memberBreakdown[m.UserID] = &response.MemberBreakdownResponse{
			UserID:     m.UserID,
			FullName:   m.FullName,
			Role:       m.Role,
			Categories: make(map[string]float64),
		}
	}
	for _, a := range aggs {
		if _, ok := memberBreakdown[a.UserID]; !ok {
			// Creator not a member anymore — include anyway
			name := ""
			role := ""
			if a.FullName.Valid {
				name = a.FullName.String
			}
			if a.Role.Valid {
				role = a.Role.String
			}
			memberBreakdown[a.UserID] = &response.MemberBreakdownResponse{
				UserID:     a.UserID,
				FullName:   name,
				Role:       role,
				Categories: make(map[string]float64),
			}
		}
		cat := normalizeCategory(a.Category)
		memberBreakdown[a.UserID].Categories[cat] += a.Amount
		memberBreakdown[a.UserID].Total += a.Amount
	}
	// Add manual expenses to breakdown
	manualExpenses, err := s.reportRepo.FindManualExpensesByProject(ctx, projectID)
	if err != nil {
		return nil, err
	}
	for _, e := range manualExpenses {
		var uid uint64
		if e.MemberUserID != nil {
			uid = *e.MemberUserID
		}
		if _, ok := memberBreakdown[uid]; !ok {
			memberBreakdown[uid] = &response.MemberBreakdownResponse{
				UserID:     uid,
				FullName:   e.MemberName,
				Categories: make(map[string]float64),
			}
		}
		cat := normalizeCategory(e.Category)
		memberBreakdown[uid].Categories[cat] += e.Amount
		memberBreakdown[uid].Total += e.Amount
	}
	for _, mb := range memberBreakdown {
		resp.MemberBreakdowns = append(resp.MemberBreakdowns, *mb)
		resp.TotalPengeluaran += mb.Total
	}

	// --- Daily expenses (per date) ---
	dailyRows, err := s.buildDailyExpenses(ctx, projectID)
	if err != nil {
		return nil, err
	}
	resp.DailyExpenses = dailyRows

	// --- Recruiter fees ---
	recruiterFees, err := s.reportRepo.FindRecruiterFeesByProject(ctx, projectID)
	if err != nil {
		return nil, err
	}
	for _, f := range recruiterFees {
		total := f.FeeRecruiter +
			(f.InsentifRespondenMain * float64(f.JumlahRespondenMain)) +
			(f.InsentifRespondenBackup * float64(f.JumlahRespondenBackup))
		resp.RecruiterFees = append(resp.RecruiterFees, response.FinanceRecruiterFeeResponse{
			ID:                      f.ID,
			RecruiterName:           f.RecruiterName,
			Jumlah:                  f.Jumlah,
			FeeRecruiter:            f.FeeRecruiter,
			InsentifRespondenMain:   f.InsentifRespondenMain,
			JumlahRespondenMain:     f.JumlahRespondenMain,
			InsentifRespondenBackup: f.InsentifRespondenBackup,
			JumlahRespondenBackup:   f.JumlahRespondenBackup,
			SortOrder:               f.SortOrder,
			Total:                   total,
			CreatedAt:               f.CreatedAt,
		})
		resp.TotalPerolehanRecruit += total
	}

	// --- Sample entries ---
	samples, err := s.reportRepo.FindSampleEntriesByProject(ctx, projectID)
	if err != nil {
		return nil, err
	}
	for _, s := range samples {
		total := (s.InsentifRespondenMain * float64(s.JumlahRespondenMain)) +
			(s.InsentifRespondenBackup * float64(s.JumlahRespondenBackup))
		resp.SampleEntries = append(resp.SampleEntries, response.FinanceSampleEntryResponse{
			ID:                      s.ID,
			TanggalPelaksanaan:      s.TanggalPelaksanaan,
			JumlahSample:            s.JumlahSample,
			InsentifRespondenMain:   s.InsentifRespondenMain,
			JumlahRespondenMain:     s.JumlahRespondenMain,
			InsentifRespondenBackup: s.InsentifRespondenBackup,
			JumlahRespondenBackup:   s.JumlahRespondenBackup,
			SortOrder:               s.SortOrder,
			Total:                   total,
		})
		resp.TotalSampleIncentive += total
		resp.JumlahMain += s.JumlahRespondenMain
		resp.JumlahBackup += s.JumlahRespondenBackup
	}

	// --- Manual expenses response ---
	for _, e := range manualExpenses {
		resp.ManualExpenses = append(resp.ManualExpenses, response.FinanceManualExpenseResponse{
			ID:           e.ID,
			MemberUserID: e.MemberUserID,
			MemberName:   e.MemberName,
			Category:     e.Category,
			Tanggal:      e.Tanggal,
			Description:  e.Description,
			Quantity:     e.Quantity,
			UnitPrice:    e.UnitPrice,
			Amount:       e.Amount,
			SortOrder:    e.SortOrder,
		})
	}

	// Grand total
	resp.TotalYangDibayarkan = resp.TotalPengeluaran + resp.TotalPerolehanRecruit + resp.TotalSampleIncentive

	return resp, nil
}

func (s *FinanceReportService) Upsert(ctx context.Context, projectID uint64, userID uint64, req *request.UpsertFinanceReportRequest) (*response.FinanceReportResponse, error) {
	// Verify project exists
	if _, err := s.projectRepo.FindByID(ctx, projectID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("project not found")
		}
		return nil, err
	}

	// Recruiter fees
	fees := make([]model.FinanceRecruiterFee, 0, len(req.RecruiterFees))
	for i, f := range req.RecruiterFees {
		sortOrder := f.SortOrder
		if sortOrder == 0 {
			sortOrder = i
		}
		fees = append(fees, model.FinanceRecruiterFee{
			RecruiterName:           f.RecruiterName,
			Jumlah:                  f.Jumlah,
			FeeRecruiter:            f.FeeRecruiter,
			InsentifRespondenMain:   f.InsentifRespondenMain,
			JumlahRespondenMain:     f.JumlahRespondenMain,
			InsentifRespondenBackup: f.InsentifRespondenBackup,
			JumlahRespondenBackup:   f.JumlahRespondenBackup,
			SortOrder:               sortOrder,
		})
	}
	if err := s.reportRepo.ReplaceRecruiterFees(ctx, projectID, fees); err != nil {
		return nil, fmt.Errorf("save recruiter fees: %w", err)
	}

	// Sample entries
	entries := make([]model.FinanceSampleEntry, 0, len(req.SampleEntries))
	for i, e := range req.SampleEntries {
		t, err := parseDateStrict(e.TanggalPelaksanaan)
		if err != nil {
			return nil, fmt.Errorf("invalid tanggal_pelaksanaan: %s", e.TanggalPelaksanaan)
		}
		sortOrder := e.SortOrder
		if sortOrder == 0 {
			sortOrder = i
		}
		entries = append(entries, model.FinanceSampleEntry{
			TanggalPelaksanaan:      t,
			JumlahSample:            e.JumlahSample,
			InsentifRespondenMain:   e.InsentifRespondenMain,
			JumlahRespondenMain:     e.JumlahRespondenMain,
			InsentifRespondenBackup: e.InsentifRespondenBackup,
			JumlahRespondenBackup:   e.JumlahRespondenBackup,
			SortOrder:               sortOrder,
		})
	}
	if err := s.reportRepo.ReplaceSampleEntries(ctx, projectID, entries); err != nil {
		return nil, fmt.Errorf("save sample entries: %w", err)
	}

	// Manual expenses
	manuals := make([]model.FinanceManualExpense, 0, len(req.ManualExpenses))
	for i, e := range req.ManualExpenses {
		var tanggal *time.Time
		if e.Tanggal != nil && *e.Tanggal != "" {
			t, err := parseDateStrict(*e.Tanggal)
			if err == nil {
				tanggal = &t
			}
		}
		amount := float64(e.Quantity) * e.UnitPrice
		if amount == 0 {
			amount = e.UnitPrice
		}
		sortOrder := e.SortOrder
		if sortOrder == 0 {
			sortOrder = i
		}
		manuals = append(manuals, model.FinanceManualExpense{
			MemberUserID: e.MemberUserID,
			MemberName:   e.MemberName,
			Category:     e.Category,
			Tanggal:      tanggal,
			Description:  e.Description,
			Quantity:     e.Quantity,
			UnitPrice:    e.UnitPrice,
			Amount:       amount,
			SortOrder:    sortOrder,
		})
	}
	if err := s.reportRepo.ReplaceManualExpenses(ctx, projectID, userID, manuals); err != nil {
		return nil, fmt.Errorf("save manual expenses: %w", err)
	}

	_, _ = s.auditRepo.Create(ctx, &model.AuditLog{
		UserID:     userID,
		Action:     "UPSERT",
		EntityType: "finance_report",
		EntityID:   projectID,
		Details:    fmt.Sprintf("recruiters=%d, samples=%d, manuals=%d", len(fees), len(entries), len(manuals)),
	})

	return s.Get(ctx, projectID)
}

// buildDailyExpenses: group expenses by DATE(created_at) per member
func (s *FinanceReportService) buildDailyExpenses(ctx context.Context, projectID uint64) ([]response.DateExpenseRow, error) {
	// Use raw query for grouping by date
	rows, err := s.reportRepo.DB().QueryContext(ctx, `
		SELECT DATE(e.created_at) as tgl, e.created_by, u.full_name, SUM(e.amount)
		FROM expenses e
		LEFT JOIN users u ON u.id = e.created_by
		WHERE e.project_id = ?
		GROUP BY DATE(e.created_at), e.created_by, u.full_name
		ORDER BY tgl ASC`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []response.DateExpenseRow
	for rows.Next() {
		var tgl time.Time
		var userID uint64
		var fullName sql.NullString
		var amount float64
		if err := rows.Scan(&tgl, &userID, &fullName, &amount); err != nil {
			return nil, err
		}
		name := ""
		if fullName.Valid {
			name = fullName.String
		}
		out = append(out, response.DateExpenseRow{
			Tanggal:    tgl.Format("2006-01-02"),
			MemberName: name,
			UangMasuk:  0,
			UangKeluar: amount,
		})
	}
	return out, rows.Err()
}

// normalizeCategory maps arbitrary category strings to normalized enum values
func normalizeCategory(raw string) string {
	up := strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(raw), " ", "_"))
	switch up {
	case "UANG_MAKAN", "UANG-MAKAN", "MAKAN":
		return model.FinCatUangMakan
	case "PULSA":
		return model.FinCatPulsa
	case "RECORDING":
		return model.FinCatRecording
	case "INPUT_PERPI", "INPUT-PERPI", "PERPI":
		return model.FinCatInputPerpi
	case "BENSIN":
		return model.FinCatBensin
	case "BRIEFING":
		return model.FinCatBriefing
	case "TRANSPORT", "TRAVEL":
		return model.FinCatTransport
	case "SPV":
		return model.FinCatSPV
	case "LAIN_LAIN", "LAINNYA", "LAIN-LAIN", "OTHER", "OTHERS":
		return model.FinCatLainLain
	default:
		return up
	}
}

func parseDateStrict(s string) (time.Time, error) {
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t, nil
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}
	return time.Time{}, fmt.Errorf("invalid date: %s", s)
}
