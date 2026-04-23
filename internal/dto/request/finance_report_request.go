package request

type FinanceRecruiterFeeRequest struct {
	RecruiterName           string  `json:"recruiter_name" validate:"required,min=1,max=255"`
	Jumlah                  int     `json:"jumlah" validate:"gte=0"`
	FeeRecruiter            float64 `json:"fee_recruiter" validate:"gte=0"`
	InsentifRespondenMain   float64 `json:"insentif_responden_main" validate:"gte=0"`
	JumlahRespondenMain     int     `json:"jumlah_responden_main" validate:"gte=0"`
	InsentifRespondenBackup float64 `json:"insentif_responden_backup" validate:"gte=0"`
	JumlahRespondenBackup   int     `json:"jumlah_responden_backup" validate:"gte=0"`
	SortOrder               int     `json:"sort_order"`
}

type FinanceSampleEntryRequest struct {
	TanggalPelaksanaan      string  `json:"tanggal_pelaksanaan" validate:"required"`
	JumlahSample            int     `json:"jumlah_sample" validate:"gte=0"`
	InsentifRespondenMain   float64 `json:"insentif_responden_main" validate:"gte=0"`
	JumlahRespondenMain     int     `json:"jumlah_responden_main" validate:"gte=0"`
	InsentifRespondenBackup float64 `json:"insentif_responden_backup" validate:"gte=0"`
	JumlahRespondenBackup   int     `json:"jumlah_responden_backup" validate:"gte=0"`
	SortOrder               int     `json:"sort_order"`
}

type FinanceManualExpenseRequest struct {
	MemberUserID *uint64 `json:"member_user_id"`
	MemberName   string  `json:"member_name" validate:"omitempty,max=255"`
	Category     string  `json:"category" validate:"required,max=100"`
	Tanggal      *string `json:"tanggal"`
	Description  string  `json:"description" validate:"omitempty,max=500"`
	Quantity     int     `json:"quantity" validate:"gte=0"`
	UnitPrice    float64 `json:"unit_price" validate:"gte=0"`
	SortOrder    int     `json:"sort_order"`
}

type UpsertFinanceReportRequest struct {
	RecruiterFees   []FinanceRecruiterFeeRequest  `json:"recruiter_fees" validate:"dive"`
	SampleEntries   []FinanceSampleEntryRequest   `json:"sample_entries" validate:"dive"`
	ManualExpenses  []FinanceManualExpenseRequest `json:"manual_expenses" validate:"dive"`
}
