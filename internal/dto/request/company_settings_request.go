package request

type UpsertCompanySettingsRequest struct {
	CompanyName       string `json:"company_name" validate:"required,max=255"`
	CompanyCode       string `json:"company_code" validate:"required,max=10"`
	Address           string `json:"address" validate:"max=1000"`
	Phone             string `json:"phone" validate:"max=50"`
	Email             string `json:"email" validate:"omitempty,email,max=255"`
	NPWP              string `json:"npwp" validate:"max=50"`
	BankName          string `json:"bank_name" validate:"max=100"`
	BankAccountNumber string `json:"bank_account_number" validate:"max=50"`
	BankAccountName   string `json:"bank_account_name" validate:"max=255"`
	BankBranch        string `json:"bank_branch" validate:"max=255"`
	LogoURL           string `json:"logo_url" validate:"omitempty,max=500"`
	SignatoryName     string `json:"signatory_name" validate:"max=255"`
	SignatoryTitle    string `json:"signatory_title" validate:"max=255"`
}
