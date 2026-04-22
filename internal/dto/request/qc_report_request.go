package request

type ApproveQCReportRequest struct {
	Notes string `json:"notes" validate:"omitempty,max=1000"`
}

type RejectQCReportRequest struct {
	Notes string `json:"notes" validate:"required,min=3,max=1000"`
}

type SubmitQCReportRequest struct{}

type QCReportItemRequest struct {
	Category  string  `json:"category" validate:"required,oneof=VISIT_URBAN VISIT_RURAL TELP_QUAL TELP_QUANT CLT_TIMESHEET RECORDING UANG_MAKAN INPUT_PERPI PARKIR BENSIN LAIN_LAIN"`
	Status    string  `json:"status" validate:"omitempty,oneof=OK DO NONE"`
	Label     string  `json:"label" validate:"omitempty,max=255"`
	Quantity  int     `json:"quantity" validate:"gte=0"`
	UnitPrice float64 `json:"unit_price" validate:"gte=0"`
	SortOrder int     `json:"sort_order"`
}

type QCRecruiterPerformanceRequest struct {
	RecruiterName string `json:"recruiter_name" validate:"required,min=1,max=255"`
	Total         int    `json:"total" validate:"gte=0"`
	OKPerpi       int    `json:"ok_perpi" validate:"gte=0"`
	DOPerpi       int    `json:"do_perpi" validate:"gte=0"`
	OKQC          int    `json:"ok_qc" validate:"gte=0"`
	DOQC          int    `json:"do_qc" validate:"gte=0"`
	Notes         string `json:"notes" validate:"omitempty,max=1000"`
	SortOrder     int    `json:"sort_order"`
}

type CreateQCReportRequest struct {
	ProjectID                 uint64                          `json:"project_id" validate:"required,gt=0"`
	QCUserID                  uint64                          `json:"qc_user_id" validate:"required,gt=0"`
	SPVNames                  string                          `json:"spv_names" validate:"omitempty,max=500"`
	ProjectType               string                          `json:"project_type" validate:"required,oneof=KUALITATIF KUANTITATIF"`
	Methodology               string                          `json:"methodology" validate:"required,oneof=FGD_TRIAD HOME_VISIT CLT IDI RANDOM"`
	City                      string                          `json:"city" validate:"omitempty,max=255"`
	Area                      string                          `json:"area" validate:"required,oneof=URBAN RURAL URBAN_RURAL"`
	ExecutionStartDate        *string                         `json:"execution_start_date"`
	ExecutionEndDate          *string                         `json:"execution_end_date"`
	BriefingDate              *string                         `json:"briefing_date"`
	WorkStartDate             *string                         `json:"work_start_date"`
	WorkEndDate               *string                         `json:"work_end_date"`
	VisitTarget               int                             `json:"visit_target" validate:"gte=0"`
	VisitOK                   int                             `json:"visit_ok" validate:"gte=0"`
	TelpTarget                int                             `json:"telp_target" validate:"gte=0"`
	TelpOK                    int                             `json:"telp_ok" validate:"gte=0"`
	Location                  string                          `json:"location" validate:"omitempty,max=255"`
	ReportDate                *string                         `json:"report_date"`
	QCSignatoryName           string                          `json:"qc_signatory_name" validate:"omitempty,max=255"`
	QCSignatoryTitle          string                          `json:"qc_signatory_title" validate:"omitempty,max=255"`
	CoordinatorSignatoryName  string                          `json:"coordinator_signatory_name" validate:"omitempty,max=255"`
	CoordinatorSignatoryTitle string                          `json:"coordinator_signatory_title" validate:"omitempty,max=255"`
	Note                      string                          `json:"note" validate:"omitempty,max=2000"`
	Items                     []QCReportItemRequest           `json:"items" validate:"dive"`
	Recruiters                []QCRecruiterPerformanceRequest `json:"recruiters" validate:"dive"`
}

type UpdateQCReportRequest struct {
	QCUserID                  uint64                          `json:"qc_user_id" validate:"omitempty,gt=0"`
	SPVNames                  string                          `json:"spv_names" validate:"omitempty,max=500"`
	ProjectType               string                          `json:"project_type" validate:"omitempty,oneof=KUALITATIF KUANTITATIF"`
	Methodology               string                          `json:"methodology" validate:"omitempty,oneof=FGD_TRIAD HOME_VISIT CLT IDI RANDOM"`
	City                      string                          `json:"city" validate:"omitempty,max=255"`
	Area                      string                          `json:"area" validate:"omitempty,oneof=URBAN RURAL URBAN_RURAL"`
	ExecutionStartDate        *string                         `json:"execution_start_date"`
	ExecutionEndDate          *string                         `json:"execution_end_date"`
	BriefingDate              *string                         `json:"briefing_date"`
	WorkStartDate             *string                         `json:"work_start_date"`
	WorkEndDate               *string                         `json:"work_end_date"`
	VisitTarget               *int                            `json:"visit_target"`
	VisitOK                   *int                            `json:"visit_ok"`
	TelpTarget                *int                            `json:"telp_target"`
	TelpOK                    *int                            `json:"telp_ok"`
	Location                  string                          `json:"location" validate:"omitempty,max=255"`
	ReportDate                *string                         `json:"report_date"`
	QCSignatoryName           string                          `json:"qc_signatory_name" validate:"omitempty,max=255"`
	QCSignatoryTitle          string                          `json:"qc_signatory_title" validate:"omitempty,max=255"`
	CoordinatorSignatoryName  string                          `json:"coordinator_signatory_name" validate:"omitempty,max=255"`
	CoordinatorSignatoryTitle string                          `json:"coordinator_signatory_title" validate:"omitempty,max=255"`
	Note                      string                          `json:"note" validate:"omitempty,max=2000"`
	Items                     []QCReportItemRequest           `json:"items" validate:"dive"`
	Recruiters                []QCRecruiterPerformanceRequest `json:"recruiters" validate:"dive"`
}
