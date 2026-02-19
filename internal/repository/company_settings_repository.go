package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/gilangrmdnii/invoice-backend/internal/model"
)

type CompanySettingsRepository struct {
	db *sql.DB
}

func NewCompanySettingsRepository(db *sql.DB) *CompanySettingsRepository {
	return &CompanySettingsRepository{db: db}
}

func (r *CompanySettingsRepository) Get(ctx context.Context) (*model.CompanySettings, error) {
	query := `SELECT id, company_name, company_code, address, phone, email, npwp,
		bank_name, bank_account_number, bank_account_name, bank_branch,
		logo_url, signatory_name, signatory_title, created_at, updated_at
	FROM company_settings LIMIT 1`

	cs := &model.CompanySettings{}
	var address, phone, email, npwp, bankName, bankAccNum, bankAccName, bankBranch sql.NullString
	var logoURL, sigName, sigTitle sql.NullString

	err := r.db.QueryRowContext(ctx, query).Scan(
		&cs.ID, &cs.CompanyName, &cs.CompanyCode, &address, &phone, &email, &npwp,
		&bankName, &bankAccNum, &bankAccName, &bankBranch,
		&logoURL, &sigName, &sigTitle, &cs.CreatedAt, &cs.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	cs.Address = address.String
	cs.Phone = phone.String
	cs.Email = email.String
	cs.NPWP = npwp.String
	cs.BankName = bankName.String
	cs.BankAccountNumber = bankAccNum.String
	cs.BankAccountName = bankAccName.String
	cs.BankBranch = bankBranch.String
	cs.LogoURL = logoURL.String
	cs.SignatoryName = sigName.String
	cs.SignatoryTitle = sigTitle.String

	return cs, nil
}

func (r *CompanySettingsRepository) Upsert(ctx context.Context, cs *model.CompanySettings) (uint64, error) {
	// Check if settings exist
	var existingID uint64
	err := r.db.QueryRowContext(ctx, `SELECT id FROM company_settings LIMIT 1`).Scan(&existingID)

	if err == sql.ErrNoRows {
		// Insert
		result, err := r.db.ExecContext(ctx,
			`INSERT INTO company_settings (company_name, company_code, address, phone, email, npwp,
				bank_name, bank_account_number, bank_account_name, bank_branch,
				logo_url, signatory_name, signatory_title)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			cs.CompanyName, cs.CompanyCode, cs.Address, cs.Phone, cs.Email, cs.NPWP,
			cs.BankName, cs.BankAccountNumber, cs.BankAccountName, cs.BankBranch,
			cs.LogoURL, cs.SignatoryName, cs.SignatoryTitle,
		)
		if err != nil {
			return 0, fmt.Errorf("insert company settings: %w", err)
		}
		id, err := result.LastInsertId()
		if err != nil {
			return 0, err
		}
		return uint64(id), nil
	}
	if err != nil {
		return 0, err
	}

	// Update
	_, err = r.db.ExecContext(ctx,
		`UPDATE company_settings SET company_name = ?, company_code = ?, address = ?,
			phone = ?, email = ?, npwp = ?, bank_name = ?, bank_account_number = ?,
			bank_account_name = ?, bank_branch = ?, logo_url = ?,
			signatory_name = ?, signatory_title = ?
		WHERE id = ?`,
		cs.CompanyName, cs.CompanyCode, cs.Address, cs.Phone, cs.Email, cs.NPWP,
		cs.BankName, cs.BankAccountNumber, cs.BankAccountName, cs.BankBranch,
		cs.LogoURL, cs.SignatoryName, cs.SignatoryTitle,
		existingID,
	)
	if err != nil {
		return 0, fmt.Errorf("update company settings: %w", err)
	}
	return existingID, nil
}
