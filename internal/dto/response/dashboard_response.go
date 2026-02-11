package response

type DashboardResponse struct {
	Projects       ProjectSummary       `json:"projects"`
	Budget         BudgetSummary        `json:"budget"`
	Expenses       ExpenseSummary       `json:"expenses"`
	BudgetRequests BudgetRequestSummary `json:"budget_requests"`
}

type ProjectSummary struct {
	TotalProjects  int64 `json:"total_projects"`
	ActiveProjects int64 `json:"active_projects"`
}

type BudgetSummary struct {
	TotalBudget float64 `json:"total_budget"`
	TotalSpent  float64 `json:"total_spent"`
	Remaining   float64 `json:"remaining"`
}

type ExpenseSummary struct {
	TotalExpenses    int64   `json:"total_expenses"`
	PendingExpenses  int64   `json:"pending_expenses"`
	ApprovedExpenses int64   `json:"approved_expenses"`
	RejectedExpenses int64   `json:"rejected_expenses"`
	TotalAmount      float64 `json:"total_amount"`
}

type BudgetRequestSummary struct {
	TotalRequests    int64   `json:"total_requests"`
	PendingRequests  int64   `json:"pending_requests"`
	ApprovedRequests int64   `json:"approved_requests"`
	RejectedRequests int64   `json:"rejected_requests"`
	TotalAmount      float64 `json:"total_amount"`
}
