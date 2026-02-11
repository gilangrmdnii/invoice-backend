package router

import (
	"database/sql"

	"github.com/gofiber/fiber/v2"
	"github.com/gilangrmdnii/invoice-backend/internal/config"
	"github.com/gilangrmdnii/invoice-backend/internal/handler"
	"github.com/gilangrmdnii/invoice-backend/internal/middleware"
	"github.com/gilangrmdnii/invoice-backend/internal/repository"
	"github.com/gilangrmdnii/invoice-backend/internal/service"
)

func SetupRoutes(app *fiber.App, db *sql.DB, cfg *config.Config) {
	// Repositories
	userRepo := repository.NewUserRepository(db)
	projectRepo := repository.NewProjectRepository(db)
	memberRepo := repository.NewProjectMemberRepository(db)
	budgetRepo := repository.NewBudgetRepository(db)
	expenseRepo := repository.NewExpenseRepository(db)
	budgetRequestRepo := repository.NewBudgetRequestRepository(db)

	// Services
	authService := service.NewAuthService(userRepo, cfg)
	projectService := service.NewProjectService(projectRepo, memberRepo, budgetRepo, userRepo)
	expenseService := service.NewExpenseService(expenseRepo, projectRepo, memberRepo)
	budgetRequestService := service.NewBudgetRequestService(budgetRequestRepo, projectRepo, memberRepo)

	// Handlers
	authHandler := handler.NewAuthHandler(authService)
	projectHandler := handler.NewProjectHandler(projectService)
	expenseHandler := handler.NewExpenseHandler(expenseService)
	budgetRequestHandler := handler.NewBudgetRequestHandler(budgetRequestService)

	api := app.Group("/api")

	// Health check
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"success": true,
			"message": "server is running",
		})
	})

	// Auth routes (public)
	auth := api.Group("/auth")
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)

	// Protected routes
	protected := api.Group("", middleware.AuthRequired(cfg.JWTSecret))

	// Project routes
	projects := protected.Group("/projects")
	projects.Post("", middleware.RequireRoles("FINANCE", "OWNER"), projectHandler.Create)
	projects.Get("", projectHandler.List)
	projects.Get("/:id", projectHandler.GetByID)
	projects.Put("/:id", middleware.RequireRoles("FINANCE", "OWNER"), projectHandler.Update)
	projects.Post("/:id/members", middleware.RequireRoles("FINANCE", "OWNER"), projectHandler.AddMember)
	projects.Delete("/:id/members/:userId", middleware.RequireRoles("FINANCE", "OWNER"), projectHandler.RemoveMember)
	projects.Get("/:id/members", projectHandler.ListMembers)

	// Expense routes
	expenses := protected.Group("/expenses")
	expenses.Post("", expenseHandler.Create)
	expenses.Get("", expenseHandler.List)
	expenses.Get("/:id", expenseHandler.GetByID)
	expenses.Put("/:id", expenseHandler.Update)
	expenses.Delete("/:id", expenseHandler.Delete)
	expenses.Post("/:id/approve", middleware.RequireRoles("FINANCE", "OWNER"), expenseHandler.Approve)
	expenses.Post("/:id/reject", middleware.RequireRoles("FINANCE", "OWNER"), expenseHandler.Reject)

	// Budget request routes
	budgetRequests := protected.Group("/budget-requests")
	budgetRequests.Post("", budgetRequestHandler.Create)
	budgetRequests.Get("", budgetRequestHandler.List)
	budgetRequests.Get("/:id", budgetRequestHandler.GetByID)
	budgetRequests.Post("/:id/approve", middleware.RequireRoles("FINANCE", "OWNER"), budgetRequestHandler.Approve)
	budgetRequests.Post("/:id/reject", middleware.RequireRoles("FINANCE", "OWNER"), budgetRequestHandler.Reject)

	// Future routes
	// api.Get("/dashboard", ...)
	// api.Get("/audit-logs", ...)
	// api.Get("/events", ...)
}
