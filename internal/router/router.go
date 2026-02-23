package router

import (
	"database/sql"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gilangrmdnii/invoice-backend/internal/config"
	"github.com/gilangrmdnii/invoice-backend/internal/handler"
	"github.com/gilangrmdnii/invoice-backend/internal/middleware"
	"github.com/gilangrmdnii/invoice-backend/internal/repository"
	"github.com/gilangrmdnii/invoice-backend/internal/service"
	"github.com/gilangrmdnii/invoice-backend/internal/sse"
)

func SetupRoutes(app *fiber.App, db *sql.DB, cfg *config.Config) {
	// SSE Hub
	sseHub := sse.NewHub()

	// Ensure uploads directory exists
	uploadDir := "./uploads"
	os.MkdirAll(uploadDir, 0755)

	// Static file serving for uploads
	app.Static("/uploads", uploadDir)

	// Repositories
	userRepo := repository.NewUserRepository(db)
	projectRepo := repository.NewProjectRepository(db)
	memberRepo := repository.NewProjectMemberRepository(db)
	budgetRepo := repository.NewBudgetRepository(db)
	expenseRepo := repository.NewExpenseRepository(db)
	budgetRequestRepo := repository.NewBudgetRequestRepository(db)
	auditLogRepo := repository.NewAuditLogRepository(db)
	notifRepo := repository.NewNotificationRepository(db)
	dashboardRepo := repository.NewDashboardRepository(db)
	invoiceRepo := repository.NewInvoiceRepository(db)
	companySettingsRepo := repository.NewCompanySettingsRepository(db)

	// Services
	authService := service.NewAuthService(userRepo, cfg)
	projectService := service.NewProjectService(projectRepo, memberRepo, budgetRepo, userRepo)
	expenseService := service.NewExpenseService(expenseRepo, projectRepo, memberRepo, auditLogRepo, notifRepo, userRepo, sseHub)
	budgetRequestService := service.NewBudgetRequestService(budgetRequestRepo, projectRepo, memberRepo, budgetRepo, auditLogRepo, notifRepo, userRepo, sseHub)
	notifService := service.NewNotificationService(notifRepo)
	dashboardService := service.NewDashboardService(dashboardRepo, projectRepo)
	invoicePaymentRepo := repository.NewInvoicePaymentRepository(db)
	invoiceService := service.NewInvoiceService(invoiceRepo, invoicePaymentRepo, projectRepo, memberRepo, auditLogRepo, notifRepo, userRepo, sseHub)
	companySettingsService := service.NewCompanySettingsService(companySettingsRepo)
	invoicePaymentService := service.NewInvoicePaymentService(invoicePaymentRepo, invoiceRepo, auditLogRepo, notifRepo, userRepo, sseHub)
	userService := service.NewUserService(userRepo, auditLogRepo)

	// Handlers
	authHandler := handler.NewAuthHandler(authService)
	projectHandler := handler.NewProjectHandler(projectService)
	expenseHandler := handler.NewExpenseHandler(expenseService)
	budgetRequestHandler := handler.NewBudgetRequestHandler(budgetRequestService)
	notifHandler := handler.NewNotificationHandler(notifService)
	dashboardHandler := handler.NewDashboardHandler(dashboardService)
	auditLogHandler := handler.NewAuditLogHandler(auditLogRepo)
	sseHandler := handler.NewSSEHandler(sseHub)
	uploadHandler := handler.NewUploadHandler(uploadDir)
	invoiceHandler := handler.NewInvoiceHandler(invoiceService)
	userHandler := handler.NewUserHandler(userService)
	companySettingsHandler := handler.NewCompanySettingsHandler(companySettingsService)
	invoicePaymentHandler := handler.NewInvoicePaymentHandler(invoicePaymentService)

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

	// File upload
	protected.Post("/upload", uploadHandler.Upload)

	// Users
	users := protected.Group("/users")
	users.Get("", userHandler.List)
	users.Post("", middleware.RequireRoles("OWNER"), userHandler.Create)
	users.Put("/:id", middleware.RequireRoles("OWNER"), userHandler.Update)
	users.Delete("/:id", middleware.RequireRoles("OWNER"), userHandler.Delete)

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

	// Invoice routes
	invoices := protected.Group("/invoices")
	invoices.Post("", invoiceHandler.Create)
	invoices.Get("", invoiceHandler.List)
	invoices.Get("/:id", invoiceHandler.GetByID)
	invoices.Put("/:id", middleware.RequireRoles("SPV"), invoiceHandler.Update)
	invoices.Delete("/:id", middleware.RequireRoles("SPV"), invoiceHandler.Delete)
	invoices.Post("/:id/approve", middleware.RequireRoles("FINANCE", "OWNER"), invoiceHandler.Approve)
	invoices.Post("/:id/reject", middleware.RequireRoles("FINANCE", "OWNER"), invoiceHandler.Reject)

	// Invoice payment routes
	invoices.Post("/:invoiceId/payments", middleware.RequireRoles("FINANCE", "OWNER"), invoicePaymentHandler.Create)
	invoices.Get("/:invoiceId/payments", invoicePaymentHandler.ListByInvoice)
	invoices.Delete("/:invoiceId/payments/:paymentId", middleware.RequireRoles("FINANCE", "OWNER"), invoicePaymentHandler.Delete)

	// Company settings (FINANCE, OWNER only)
	companySettings := protected.Group("/company-settings")
	companySettings.Get("", companySettingsHandler.Get)
	companySettings.Put("", middleware.RequireRoles("FINANCE", "OWNER"), companySettingsHandler.Upsert)

	// Dashboard
	protected.Get("/dashboard", dashboardHandler.GetDashboard)

	// Notifications
	notifications := protected.Group("/notifications")
	notifications.Get("", notifHandler.List)
	notifications.Get("/unread-count", notifHandler.CountUnread)
	notifications.Patch("/read-all", notifHandler.MarkAllAsRead)
	notifications.Patch("/:id/read", notifHandler.MarkAsRead)

	// Audit logs (FINANCE, OWNER only)
	protected.Get("/audit-logs", middleware.RequireRoles("FINANCE", "OWNER"), auditLogHandler.List)

	// SSE events
	protected.Get("/events", sseHandler.Stream)
}
