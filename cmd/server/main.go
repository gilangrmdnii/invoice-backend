package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gilangrmdnii/invoice-backend/internal/config"
	"github.com/gilangrmdnii/invoice-backend/internal/database"
	"github.com/gilangrmdnii/invoice-backend/internal/router"
	"github.com/gilangrmdnii/invoice-backend/migrations"
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

	// Run migrations if AUTO_MIGRATE=true
	if os.Getenv("AUTO_MIGRATE") == "true" {
		log.Println("running auto-migration...")
		if err := database.RunMigrations(db, migrations.Files); err != nil {
			log.Fatalf("migration failed: %v", err)
		}
		log.Println("migrations completed")
	}

	app := fiber.New(fiber.Config{
		AppName: "Invoice Backend",
	})

	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New())

	router.SetupRoutes(app, db, cfg)

	port := os.Getenv("PORT")
	if port == "" {
		port = cfg.AppPort
	}

	log.Printf("server starting on port %s", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
