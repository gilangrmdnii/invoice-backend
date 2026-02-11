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

type ExpenseService struct {
	expenseRepo *repository.ExpenseRepository
	projectRepo *repository.ProjectRepository
	memberRepo  *repository.ProjectMemberRepository
}

func NewExpenseService(
	expenseRepo *repository.ExpenseRepository,
	projectRepo *repository.ProjectRepository,
	memberRepo *repository.ProjectMemberRepository,
) *ExpenseService {
	return &ExpenseService{
		expenseRepo: expenseRepo,
		projectRepo: projectRepo,
		memberRepo:  memberRepo,
	}
}

func (s *ExpenseService) Create(ctx context.Context, req *request.CreateExpenseRequest, userID uint64, role string) (*response.ExpenseResponse, error) {
	// Verify project exists
	_, err := s.projectRepo.FindByID(ctx, req.ProjectID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("project not found")
		}
		return nil, err
	}

	// SPV must be a member of the project
	if role == string(model.RoleSPV) {
		isMember, err := s.memberRepo.Exists(ctx, req.ProjectID, userID)
		if err != nil {
			return nil, err
		}
		if !isMember {
			return nil, fmt.Errorf("not a member of this project")
		}
	}

	expense := &model.Expense{
		ProjectID:   req.ProjectID,
		Description: req.Description,
		Amount:      req.Amount,
		Category:    req.Category,
		ReceiptURL:  req.ReceiptURL,
		Status:      model.ExpenseStatusPending,
		CreatedBy:   userID,
	}

	id, err := s.expenseRepo.Create(ctx, expense)
	if err != nil {
		return nil, fmt.Errorf("create expense: %w", err)
	}

	return &response.ExpenseResponse{
		ID:          id,
		ProjectID:   expense.ProjectID,
		Description: expense.Description,
		Amount:      expense.Amount,
		Category:    expense.Category,
		ReceiptURL:  expense.ReceiptURL,
		Status:      string(expense.Status),
		CreatedBy:   userID,
	}, nil
}

func (s *ExpenseService) List(ctx context.Context, userID uint64, role string) ([]response.ExpenseResponse, error) {
	var expenses []model.Expense
	var err error

	if role == string(model.RoleSPV) {
		// SPV sees only expenses from their projects
		projects, err := s.projectRepo.FindByMemberUserID(ctx, userID)
		if err != nil {
			return nil, err
		}
		projectIDs := make([]uint64, len(projects))
		for i, p := range projects {
			projectIDs[i] = p.ID
		}
		expenses, err = s.expenseRepo.FindByProjectIDs(ctx, projectIDs)
		if err != nil {
			return nil, err
		}
	} else {
		expenses, err = s.expenseRepo.FindAll(ctx)
	}
	if err != nil {
		return nil, err
	}

	result := make([]response.ExpenseResponse, 0, len(expenses))
	for _, e := range expenses {
		result = append(result, toExpenseResponse(&e))
	}
	return result, nil
}

func (s *ExpenseService) GetByID(ctx context.Context, id uint64) (*response.ExpenseResponse, error) {
	expense, err := s.expenseRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("expense not found")
		}
		return nil, err
	}
	resp := toExpenseResponse(expense)
	return &resp, nil
}

func (s *ExpenseService) Update(ctx context.Context, id uint64, req *request.UpdateExpenseRequest, userID uint64, role string) (*response.ExpenseResponse, error) {
	expense, err := s.expenseRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("expense not found")
		}
		return nil, err
	}

	// SPV can only update their own expenses
	if role == string(model.RoleSPV) && expense.CreatedBy != userID {
		return nil, fmt.Errorf("not authorized to update this expense")
	}

	// Can only update PENDING expenses
	if expense.Status != model.ExpenseStatusPending {
		return nil, fmt.Errorf("only pending expenses can be updated")
	}

	if req.Description != "" {
		expense.Description = req.Description
	}
	if req.Amount > 0 {
		expense.Amount = req.Amount
	}
	if req.Category != "" {
		expense.Category = req.Category
	}
	if req.ReceiptURL != "" {
		expense.ReceiptURL = req.ReceiptURL
	}

	if err := s.expenseRepo.Update(ctx, expense); err != nil {
		return nil, fmt.Errorf("update expense: %w", err)
	}

	updated, err := s.expenseRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	resp := toExpenseResponse(updated)
	return &resp, nil
}

func (s *ExpenseService) Delete(ctx context.Context, id uint64, userID uint64, role string) error {
	expense, err := s.expenseRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("expense not found")
		}
		return err
	}

	// SPV can only delete their own expenses
	if role == string(model.RoleSPV) && expense.CreatedBy != userID {
		return fmt.Errorf("not authorized to delete this expense")
	}

	// Can only delete PENDING expenses
	if expense.Status != model.ExpenseStatusPending {
		return fmt.Errorf("only pending expenses can be deleted")
	}

	return s.expenseRepo.Delete(ctx, id)
}

func (s *ExpenseService) Approve(ctx context.Context, id uint64, approvedBy uint64, notes string) (*response.ExpenseResponse, error) {
	if err := s.expenseRepo.ApproveExpense(ctx, id, approvedBy, notes); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("expense not found")
		}
		if err.Error() == "expense is not pending" {
			return nil, err
		}
		return nil, fmt.Errorf("approve expense: %w", err)
	}

	expense, err := s.expenseRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	resp := toExpenseResponse(expense)
	return &resp, nil
}

func (s *ExpenseService) Reject(ctx context.Context, id uint64, approvedBy uint64, notes string) (*response.ExpenseResponse, error) {
	if err := s.expenseRepo.RejectExpense(ctx, id, approvedBy, notes); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("expense not found")
		}
		if err.Error() == "expense is not pending" {
			return nil, err
		}
		return nil, fmt.Errorf("reject expense: %w", err)
	}

	expense, err := s.expenseRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	resp := toExpenseResponse(expense)
	return &resp, nil
}

func toExpenseResponse(e *model.Expense) response.ExpenseResponse {
	return response.ExpenseResponse{
		ID:          e.ID,
		ProjectID:   e.ProjectID,
		Description: e.Description,
		Amount:      e.Amount,
		Category:    e.Category,
		ReceiptURL:  e.ReceiptURL,
		Status:      string(e.Status),
		CreatedBy:   e.CreatedBy,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}
}
