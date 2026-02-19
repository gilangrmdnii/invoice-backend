package response

import "time"

type CompanySettingsResponse struct {
	ID                uint64    `json:"id"`
	CompanyName       string    `json:"company_name"`
	CompanyCode       string    `json:"company_code"`
	Address           string    `json:"address,omitempty"`
	Phone             string    `json:"phone,omitempty"`
	Email             string    `json:"email,omitempty"`
	NPWP              string    `json:"npwp,omitempty"`
	BankName          string    `json:"bank_name,omitempty"`
	BankAccountNumber string    `json:"bank_account_number,omitempty"`
	BankAccountName   string    `json:"bank_account_name,omitempty"`
	BankBranch        string    `json:"bank_branch,omitempty"`
	LogoURL           string    `json:"logo_url,omitempty"`
	SignatoryName     string    `json:"signatory_name,omitempty"`
	SignatoryTitle    string    `json:"signatory_title,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}
