package model

import "time"

type FinanceRecruiterFee struct {
	ID                      uint64    `json:"id"`
	ProjectID               uint64    `json:"project_id"`
	RecruiterName           string    `json:"recruiter_name"`
	Jumlah                  int       `json:"jumlah"`
	FeeRecruiter            float64   `json:"fee_recruiter"`
	InsentifRespondenMain   float64   `json:"insentif_responden_main"`
	JumlahRespondenMain     int       `json:"jumlah_responden_main"`
	InsentifRespondenBackup float64   `json:"insentif_responden_backup"`
	JumlahRespondenBackup   int       `json:"jumlah_responden_backup"`
	SortOrder               int       `json:"sort_order"`
	CreatedAt               time.Time `json:"created_at"`
	UpdatedAt               time.Time `json:"updated_at"`
}

type FinanceSampleEntry struct {
	ID                      uint64    `json:"id"`
	ProjectID               uint64    `json:"project_id"`
	TanggalPelaksanaan      time.Time `json:"tanggal_pelaksanaan"`
	JumlahSample            int       `json:"jumlah_sample"`
	InsentifRespondenMain   float64   `json:"insentif_responden_main"`
	JumlahRespondenMain     int       `json:"jumlah_responden_main"`
	InsentifRespondenBackup float64   `json:"insentif_responden_backup"`
	JumlahRespondenBackup   int       `json:"jumlah_responden_backup"`
	SortOrder               int       `json:"sort_order"`
	CreatedAt               time.Time `json:"created_at"`
	UpdatedAt               time.Time `json:"updated_at"`
}

type FinanceManualExpense struct {
	ID           uint64     `json:"id"`
	ProjectID    uint64     `json:"project_id"`
	MemberUserID *uint64    `json:"member_user_id"`
	MemberName   string     `json:"member_name"`
	Category     string     `json:"category"`
	Tanggal      *time.Time `json:"tanggal"`
	Description  string     `json:"description"`
	Quantity     int        `json:"quantity"`
	UnitPrice    float64    `json:"unit_price"`
	Amount       float64    `json:"amount"`
	SortOrder    int        `json:"sort_order"`
	CreatedBy    uint64     `json:"created_by"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// Aggregation category constants for Finance Report breakdown.
// Maps to both Expense.Category values AND FinanceManualExpense.Category values.
const (
	FinCatSPV         = "SPV"
	FinCatUangMakan   = "UANG_MAKAN"
	FinCatPulsa       = "PULSA"
	FinCatRecording   = "RECORDING"
	FinCatInputPerpi  = "INPUT_PERPI"
	FinCatBensin      = "BENSIN"
	FinCatBriefing    = "BRIEFING"
	FinCatTransport   = "TRANSPORT"
	FinCatLainLain    = "LAIN_LAIN"
)
