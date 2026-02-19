package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/gilangrmdnii/invoice-backend/internal/dto/request"
	"github.com/gilangrmdnii/invoice-backend/internal/dto/response"
	"github.com/gilangrmdnii/invoice-backend/internal/model"
	"github.com/gilangrmdnii/invoice-backend/internal/repository"
)

type CompanySettingsService struct {
	repo *repository.CompanySettingsRepository
}

func NewCompanySettingsService(repo *repository.CompanySettingsRepository) *CompanySettingsService {
	return &CompanySettingsService{repo: repo}
}

func (s *CompanySettingsService) Get(ctx context.Context) (*response.CompanySettingsResponse, error) {
	cs, err := s.repo.Get(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("company settings not found")
		}
		return nil, err
	}
	return toCompanySettingsResponse(cs), nil
}

func (s *CompanySettingsService) Upsert(ctx context.Context, req *request.UpsertCompanySettingsRequest) (*response.CompanySettingsResponse, error) {
	cs := &model.CompanySettings{
		CompanyName:       req.CompanyName,
		CompanyCode:       req.CompanyCode,
		Address:           req.Address,
		Phone:             req.Phone,
		Email:             req.Email,
		NPWP:              req.NPWP,
		BankName:          req.BankName,
		BankAccountNumber: req.BankAccountNumber,
		BankAccountName:   req.BankAccountName,
		BankBranch:        req.BankBranch,
		LogoURL:           req.LogoURL,
		SignatoryName:     req.SignatoryName,
		SignatoryTitle:    req.SignatoryTitle,
	}

	_, err := s.repo.Upsert(ctx, cs)
	if err != nil {
		return nil, fmt.Errorf("upsert company settings: %w", err)
	}

	return s.Get(ctx)
}

func toCompanySettingsResponse(cs *model.CompanySettings) *response.CompanySettingsResponse {
	return &response.CompanySettingsResponse{
		ID:                cs.ID,
		CompanyName:       cs.CompanyName,
		CompanyCode:       cs.CompanyCode,
		Address:           cs.Address,
		Phone:             cs.Phone,
		Email:             cs.Email,
		NPWP:              cs.NPWP,
		BankName:          cs.BankName,
		BankAccountNumber: cs.BankAccountNumber,
		BankAccountName:   cs.BankAccountName,
		BankBranch:        cs.BankBranch,
		LogoURL:           cs.LogoURL,
		SignatoryName:     cs.SignatoryName,
		SignatoryTitle:    cs.SignatoryTitle,
		CreatedAt:         cs.CreatedAt,
		UpdatedAt:         cs.UpdatedAt,
	}
}
