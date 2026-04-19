package model

import "time"

type QCProjectType string

const (
	QCProjectKualitatif  QCProjectType = "KUALITATIF"
	QCProjectKuantitatif QCProjectType = "KUANTITATIF"
)

type QCMethodology string

const (
	QCMethodFGDTriad  QCMethodology = "FGD_TRIAD"
	QCMethodHomeVisit QCMethodology = "HOME_VISIT"
	QCMethodCLT       QCMethodology = "CLT"
	QCMethodIDI       QCMethodology = "IDI"
	QCMethodRandom    QCMethodology = "RANDOM"
)

type QCArea string

const (
	QCAreaUrban      QCArea = "URBAN"
	QCAreaRural      QCArea = "RURAL"
	QCAreaUrbanRural QCArea = "URBAN_RURAL"
)

type QCItemCategory string

const (
	QCCatVisitUrban   QCItemCategory = "VISIT_URBAN"
	QCCatVisitRural   QCItemCategory = "VISIT_RURAL"
	QCCatTelpQual     QCItemCategory = "TELP_QUAL"
	QCCatTelpQuant    QCItemCategory = "TELP_QUANT"
	QCCatCLTTimeSheet QCItemCategory = "CLT_TIMESHEET"
	QCCatRecording    QCItemCategory = "RECORDING"
	QCCatUangMakan    QCItemCategory = "UANG_MAKAN"
	QCCatInputPerpi   QCItemCategory = "INPUT_PERPI"
	QCCatParkir       QCItemCategory = "PARKIR"
	QCCatBensin       QCItemCategory = "BENSIN"
	QCCatLainLain     QCItemCategory = "LAIN_LAIN"
)

type QCItemStatus string

const (
	QCStatusOK   QCItemStatus = "OK"
	QCStatusDO   QCItemStatus = "DO"
	QCStatusNone QCItemStatus = "NONE"
)

type QCReport struct {
	ID                        uint64        `json:"id"`
	ProjectID                 uint64        `json:"project_id"`
	QCUserID                  uint64        `json:"qc_user_id"`
	SPVNames                  string        `json:"spv_names"`
	ProjectType               QCProjectType `json:"project_type"`
	Methodology               QCMethodology `json:"methodology"`
	City                      string        `json:"city"`
	Area                      QCArea        `json:"area"`
	ExecutionStartDate        *time.Time    `json:"execution_start_date"`
	ExecutionEndDate          *time.Time    `json:"execution_end_date"`
	BriefingDate              *time.Time    `json:"briefing_date"`
	WorkStartDate             *time.Time    `json:"work_start_date"`
	WorkEndDate               *time.Time    `json:"work_end_date"`
	VisitTarget               int           `json:"visit_target"`
	VisitOK                   int           `json:"visit_ok"`
	TelpTarget                int           `json:"telp_target"`
	TelpOK                    int           `json:"telp_ok"`
	TotalAmount               float64       `json:"total_amount"`
	Location                  string        `json:"location"`
	ReportDate                *time.Time    `json:"report_date"`
	QCSignatoryName           string        `json:"qc_signatory_name"`
	QCSignatoryTitle          string        `json:"qc_signatory_title"`
	CoordinatorSignatoryName  string        `json:"coordinator_signatory_name"`
	CoordinatorSignatoryTitle string        `json:"coordinator_signatory_title"`
	Note                      string        `json:"note"`
	CreatedBy                 uint64        `json:"created_by"`
	CreatedAt                 time.Time     `json:"created_at"`
	UpdatedAt                 time.Time     `json:"updated_at"`
}

type QCReportItem struct {
	ID         uint64         `json:"id"`
	QCReportID uint64         `json:"qc_report_id"`
	Category   QCItemCategory `json:"category"`
	Status     QCItemStatus   `json:"status"`
	Label      string         `json:"label"`
	Quantity   int            `json:"quantity"`
	UnitPrice  float64        `json:"unit_price"`
	Subtotal   float64        `json:"subtotal"`
	SortOrder  int            `json:"sort_order"`
	CreatedAt  time.Time      `json:"created_at"`
}

type QCRecruiterPerformance struct {
	ID            uint64    `json:"id"`
	QCReportID    uint64    `json:"qc_report_id"`
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
