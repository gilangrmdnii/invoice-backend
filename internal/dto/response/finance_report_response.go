package response

import "time"

type FinanceRecruiterFeeResponse struct {
	ID                      uint64    `json:"id"`
	RecruiterName           string    `json:"recruiter_name"`
	Jumlah                  int       `json:"jumlah"`
	FeeRecruiter            float64   `json:"fee_recruiter"`
	InsentifRespondenMain   float64   `json:"insentif_responden_main"`
	JumlahRespondenMain     int       `json:"jumlah_responden_main"`
	InsentifRespondenBackup float64   `json:"insentif_responden_backup"`
	JumlahRespondenBackup   int       `json:"jumlah_responden_backup"`
	SortOrder               int       `json:"sort_order"`
	Total                   float64   `json:"total"` // fee + insentif*jumlah main/backup
	CreatedAt               time.Time `json:"created_at"`
}

type FinanceSampleEntryResponse struct {
	ID                      uint64    `json:"id"`
	TanggalPelaksanaan      time.Time `json:"tanggal_pelaksanaan"`
	JumlahSample            int       `json:"jumlah_sample"`
	InsentifRespondenMain   float64   `json:"insentif_responden_main"`
	JumlahRespondenMain     int       `json:"jumlah_responden_main"`
	InsentifRespondenBackup float64   `json:"insentif_responden_backup"`
	JumlahRespondenBackup   int       `json:"jumlah_responden_backup"`
	SortOrder               int       `json:"sort_order"`
	Total                   float64   `json:"total"`
}

type FinanceManualExpenseResponse struct {
	ID           uint64     `json:"id"`
	MemberUserID *uint64    `json:"member_user_id"`
	MemberName   string     `json:"member_name"`
	Category     string     `json:"category"`
	Tanggal      *time.Time `json:"tanggal"`
	Description  string     `json:"description"`
	Quantity     int        `json:"quantity"`
	UnitPrice    float64    `json:"unit_price"`
	Amount       float64    `json:"amount"`
	SortOrder    int        `json:"sort_order"`
}

// MemberBreakdown: one row per project member (SPV/QC) with category totals
type MemberBreakdownResponse struct {
	UserID     uint64             `json:"user_id"`
	FullName   string             `json:"full_name"`
	Role       string             `json:"role"`
	Categories map[string]float64 `json:"categories"` // key=category, value=amount
	Total      float64            `json:"total"`
}

// DateExpenseSummary: uang masuk/keluar per tanggal per member (dari expenses.created_at date + expenses.amount)
type DateExpenseRow struct {
	Tanggal    string  `json:"tanggal"`    // "YYYY-MM-DD"
	MemberName string  `json:"member_name"`
	UangMasuk  float64 `json:"uang_masuk"`
	UangKeluar float64 `json:"uang_keluar"`
}

type FinanceReportResponse struct {
	ProjectID          uint64                         `json:"project_id"`
	ProjectName        string                         `json:"project_name"`
	ExecutionStartDate *time.Time                     `json:"execution_start_date"`
	ExecutionEndDate   *time.Time                     `json:"execution_end_date"`
	SPVNames           string                         `json:"spv_names"`
	QCNames            string                         `json:"qc_names"`
	JumlahMain         int                            `json:"jumlah_main"`
	JumlahBackup       int                            `json:"jumlah_backup"`

	// Aggregated breakdown per member
	MemberBreakdowns []MemberBreakdownResponse `json:"member_breakdowns"`

	// Daily expense log (from expenses table)
	DailyExpenses []DateExpenseRow `json:"daily_expenses"`

	// Editable sections
	RecruiterFees  []FinanceRecruiterFeeResponse  `json:"recruiter_fees"`
	SampleEntries  []FinanceSampleEntryResponse   `json:"sample_entries"`
	ManualExpenses []FinanceManualExpenseResponse `json:"manual_expenses"`

	// Grand totals
	TotalPengeluaran       float64 `json:"total_pengeluaran"`       // sum all expenses + manual
	TotalPerolehanRecruit  float64 `json:"total_perolehan_recruit"` // sum recruiter fees totals
	TotalSampleIncentive   float64 `json:"total_sample_incentive"`  // sum sample entries totals
	TotalYangDibayarkan    float64 `json:"total_yang_dibayarkan"`   // total semua yang harus dibayar
}
