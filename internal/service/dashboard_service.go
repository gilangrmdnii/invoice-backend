package service

import (
	"context"

	"github.com/gilangrmdnii/invoice-backend/internal/dto/response"
	"github.com/gilangrmdnii/invoice-backend/internal/model"
	"github.com/gilangrmdnii/invoice-backend/internal/repository"
)

type DashboardService struct {
	dashboardRepo *repository.DashboardRepository
	projectRepo   *repository.ProjectRepository
}

func NewDashboardService(
	dashboardRepo *repository.DashboardRepository,
	projectRepo *repository.ProjectRepository,
) *DashboardService {
	return &DashboardService{
		dashboardRepo: dashboardRepo,
		projectRepo:   projectRepo,
	}
}

func (s *DashboardService) GetDashboard(ctx context.Context, userID uint64, role string) (*response.DashboardResponse, error) {
	var projectIDs []uint64

	// SPV only sees their own projects
	if role == string(model.RoleSPV) {
		projects, err := s.projectRepo.FindByMemberUserID(ctx, userID)
		if err != nil {
			return nil, err
		}
		projectIDs = make([]uint64, len(projects))
		for i, p := range projects {
			projectIDs[i] = p.ID
		}
	}

	projectSummary, err := s.dashboardRepo.GetProjectSummary(ctx, projectIDs)
	if err != nil {
		return nil, err
	}

	budgetSummary, err := s.dashboardRepo.GetBudgetSummary(ctx, projectIDs)
	if err != nil {
		return nil, err
	}

	expenseSummary, err := s.dashboardRepo.GetExpenseSummary(ctx, projectIDs)
	if err != nil {
		return nil, err
	}

	budgetRequestSummary, err := s.dashboardRepo.GetBudgetRequestSummary(ctx, projectIDs)
	if err != nil {
		return nil, err
	}

	invoiceSummary, err := s.dashboardRepo.GetInvoiceSummary(ctx, projectIDs)
	if err != nil {
		return nil, err
	}

	return &response.DashboardResponse{
		Projects: response.ProjectSummary{
			TotalProjects:  projectSummary.TotalProjects,
			ActiveProjects: projectSummary.ActiveProjects,
		},
		Budget: response.BudgetSummary{
			TotalBudget: budgetSummary.TotalBudget,
			TotalSpent:  budgetSummary.TotalSpent,
			Remaining:   budgetSummary.Remaining,
		},
		Expenses: response.ExpenseSummary{
			TotalExpenses:    expenseSummary.TotalExpenses,
			PendingExpenses:  expenseSummary.PendingExpenses,
			ApprovedExpenses: expenseSummary.ApprovedExpenses,
			RejectedExpenses: expenseSummary.RejectedExpenses,
			TotalAmount:      expenseSummary.TotalAmount,
		},
		BudgetRequests: response.BudgetRequestSummary{
			TotalRequests:    budgetRequestSummary.TotalRequests,
			PendingRequests:  budgetRequestSummary.PendingRequests,
			ApprovedRequests: budgetRequestSummary.ApprovedRequests,
			RejectedRequests: budgetRequestSummary.RejectedRequests,
			TotalAmount:      budgetRequestSummary.TotalAmount,
		},
		Invoices: response.InvoiceSummary{
			TotalInvoices: invoiceSummary.TotalInvoices,
			TotalAmount:   invoiceSummary.TotalAmount,
		},
	}, nil
}
