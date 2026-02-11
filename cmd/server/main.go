package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gilangrmdnii/invoice-backend/internal/config"
	"github.com/gilangrmdnii/invoice-backend/internal/database"
	"github.com/gilangrmdnii/invoice-backend/internal/router"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	db, err := database.NewMySQL(cfg)
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	defer db.Close()
	log.Println("database connected successfully")

	app := fiber.New(fiber.Config{
		AppName: "Invoice Backend",
	})

	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New())

	router.SetupRoutes(app, db, cfg)

	log.Printf("server starting on port %s", cfg.AppPort)
	if err := app.Listen(":" + cfg.AppPort); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
