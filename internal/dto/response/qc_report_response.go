package response

import "time"

type QCReportItemResponse struct {
	ID        uint64    `json:"id"`
	Category  string    `json:"category"`
	Status    string    `json:"status"`
	Label     string    `json:"label"`
	Quantity  int       `json:"quantity"`
	UnitPrice float64   `json:"unit_price"`
	Subtotal  float64   `json:"subtotal"`
	SortOrder int       `json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
}

type QCRecruiterPerformanceResponse struct {
	ID            uint64    `json:"id"`
	RecruiterName string    `json:"recruiter_name"`
	Total         int       `json:"total"`
	OKPerpi       int       `json:"ok_perpi"`
	DOPerpi       int       `json:"do_perpi"`
	OKQC          int       `json:"ok_qc"`
	DOQC          int       `json:"do_qc"`
	Notes         string    `json:"notes"`
	SortOrder     int       `json:"sort_order"`
	CreatedAt     time.Time `json:"created_at"`
}

type QCReportResponse struct {
	ID                        uint64                           `json:"id"`
	ProjectID                 uint64                           `json:"project_id"`
	ProjectName               string                           `json:"project_name"`
	QCUserID                  uint64                           `json:"qc_user_id"`
	QCUserName                string                           `json:"qc_user_name"`
	SPVNames                  string                           `json:"spv_names"`
	ProjectType               string                           `json:"project_type"`
	Methodology               string                           `json:"methodology"`
	City                      string                           `json:"city"`
	Area                      string                           `json:"area"`
	ExecutionStartDate        *time.Time                       `json:"execution_start_date"`
	ExecutionEndDate          *time.Time                       `json:"execution_end_date"`
	BriefingDate              *time.Time                       `json:"briefing_date"`
	WorkStartDate             *time.Time                       `json:"work_start_date"`
	WorkEndDate               *time.Time                       `json:"work_end_date"`
	VisitTarget               int                              `json:"visit_target"`
	VisitOK                   int                              `json:"visit_ok"`
	TelpTarget                int                              `json:"telp_target"`
	TelpOK                    int                              `json:"telp_ok"`
	TotalAmount               float64                          `json:"total_amount"`
	Location                  string                           `json:"location"`
	ReportDate                *time.Time                       `json:"report_date"`
	QCSignatoryName           string                           `json:"qc_signatory_name"`
	QCSignatoryTitle          string                           `json:"qc_signatory_title"`
	CoordinatorSignatoryName  string                           `json:"coordinator_signatory_name"`
	CoordinatorSignatoryTitle string                           `json:"coordinator_signatory_title"`
	Note                      string                           `json:"note"`
	CreatedBy                 uint64                           `json:"created_by"`
	CreatorName               string                           `json:"creator_name"`
	Items                     []QCReportItemResponse           `json:"items"`
	Recruiters                []QCRecruiterPerformanceResponse `json:"recruiters"`
	CreatedAt                 time.Time                        `json:"created_at"`
	UpdatedAt                 time.Time                        `json:"updated_at"`
}
