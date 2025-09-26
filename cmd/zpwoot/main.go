// Package main provides the entry point for zpwoot application
//
// @title zpwoot - WhatsApp Multi-Session API
// @version 1.0
// @description A complete REST API for managing multiple WhatsApp sessions using Go, Fiber, PostgreSQL, and whatsmeow library.
// @description
// @description ## Authentication
// @description All API endpoints (except /health and /swagger/*) require API key authentication.
// @description Provide your API key in the `Authorization` header.
//
// @contact.name zpwoot Support
// @contact.url https://github.com/your-org/zpwoot
// @contact.email support@zpwoot.com
//
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
//
// @host localhost:8080
// @BasePath /
//
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
// @description Enter your API key directly (no Bearer prefix required). Example: dev-api-key-12345
package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"

	_ "zpwoot/docs/swagger" // Import generated swagger docs
	"zpwoot/internal/app"
	"zpwoot/internal/infra/db"
	"zpwoot/internal/infra/http/middleware"
	"zpwoot/internal/infra/http/routers"
	"zpwoot/internal/infra/repository"
	"zpwoot/internal/infra/wmeow"
	"zpwoot/platform/config"
	platformDB "zpwoot/platform/db"
	"zpwoot/platform/logger"
)

// Build information - set via ldflags during build
var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

func main() {
	// Define command line flags
	var (
		migrateUp     = flag.Bool("migrate-up", false, "Run database migrations up")
		migrateDown   = flag.Bool("migrate-down", false, "Rollback last migration")
		migrateStatus = flag.Bool("migrate-status", false, "Show migration status")
		seed          = flag.Bool("seed", false, "Seed database with sample data")
		version       = flag.Bool("version", false, "Show version information")
	)
	flag.Parse()

	// Handle version flag
	if *version {
		showVersion()
		return
	}

	// Load configuration
	cfg := config.Load()

	// Initialize logger with configuration based on environment
	loggerConfig := &logger.LogConfig{
		Level:      cfg.LogLevel,
		Format:     cfg.LogFormat,
		Output:     cfg.LogOutput,
		TimeFormat: "2006-01-02 15:04:05",
		Caller:     cfg.IsDevelopment(), // Show caller info in development
	}

	// Use production config if in production
	if cfg.IsProduction() {
		loggerConfig = logger.ProductionConfig()
		loggerConfig.Level = cfg.LogLevel // Override with env setting
	}

	appLogger := logger.NewWithConfig(loggerConfig)

	// Initialize database with automatic migrations
	database, err := platformDB.NewWithMigrations(cfg.DatabaseURL, appLogger)
	if err != nil {
		appLogger.Fatal("Failed to connect to database and run migrations: " + err.Error())
	}
	defer func() {
		if err := database.Close(); err != nil {
			appLogger.Error("Failed to close database connection: " + err.Error())
		}
	}()

	// Handle migration and seeding flags
	migrator := db.NewMigrator(database.GetDB().DB, appLogger)

	if *migrateUp {
		if err := migrator.RunMigrations(); err != nil {
			appLogger.Fatal("Failed to run migrations: " + err.Error())
		}
		appLogger.Info("Migrations completed successfully")
		return
	}

	if *migrateDown {
		if err := migrator.Rollback(); err != nil {
			appLogger.Fatal("Failed to rollback migration: " + err.Error())
		}
		appLogger.Info("Migration rollback completed successfully")
		return
	}

	if *migrateStatus {
		migrations, err := migrator.GetMigrationStatus()
		if err != nil {
			appLogger.Fatal("Failed to get migration status: " + err.Error())
		}
		showMigrationStatus(migrations, appLogger)
		return
	}

	if *seed {
		if err := seedDatabase(database, appLogger); err != nil {
			appLogger.Fatal("Failed to seed database: " + err.Error())
		}
		appLogger.Info("Database seeding completed successfully")
		return
	}

	// Initialize repositories
	repositories := repository.NewRepositories(database.GetDB(), appLogger)

	// Initialize WhatsApp manager and create whatsmeow tables
	appLogger.Info("Initializing WhatsApp manager and creating whatsmeow tables...")
	whatsappManager, err := initializeWhatsAppManager(database, appLogger)
	if err != nil {
		appLogger.Fatal("Failed to initialize WhatsApp manager: " + err.Error())
	}
	appLogger.Info("WhatsApp manager initialized successfully with whatsmeow tables created")

	// Initialize application container with dependencies
	container := app.NewContainer(&app.ContainerConfig{
		SessionRepo:         repositories.GetSessionRepository(),
		WebhookRepo:         repositories.GetWebhookRepository(),
		ChatwootRepo:        repositories.GetChatwootRepository(),
		WhatsAppManager:     whatsappManager,
		ChatwootIntegration: nil, // Will be implemented when Chatwoot integration is needed
		Logger:              appLogger,
		DB:                  database.GetDB().DB,
		Version:             Version,
		BuildTime:           BuildTime,
		GitCommit:           GitCommit,
	})

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true, // Disable the Fiber startup banner
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// Middleware
	app.Use(recover.New())
	app.Use(middleware.RequestID(appLogger))
	app.Use(middleware.HTTPLogger(appLogger))
	app.Use(middleware.Metrics(container, appLogger))
	app.Use(cors.New())
	app.Use(middleware.APIKeyAuth(cfg, appLogger))

	// Setup routes with dependencies
	routers.SetupRoutes(app, database, appLogger, whatsappManager, container)

	// Graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		appLogger.Info("Shutting down server...")
		if err := app.Shutdown(); err != nil {
			appLogger.Error("Failed to shutdown server gracefully: " + err.Error())
		}
	}()

	// Start server
	appLogger.InfoWithFields("Starting zpwoot server", map[string]interface{}{
		"port":        cfg.Port,
		"server_host": cfg.ServerHost,
		"environment": cfg.NodeEnv,
		"log_level":   cfg.LogLevel,
	})
	if err := app.Listen(":" + cfg.Port); err != nil {
		appLogger.Fatal("Server failed to start: " + err.Error())
	}
}

// initializeWhatsAppManager creates and initializes the WhatsApp manager
// This will automatically create the whatsmeow tables in the database
func initializeWhatsAppManager(database *db.DB, appLogger *logger.Logger) (*wmeow.Manager, error) {
	appLogger.Info("Creating WhatsApp manager factory...")

	// For now, we'll pass nil for sessionRepo since it's not fully implemented yet
	// In a complete implementation, you would initialize the session repository here
	factory := wmeow.NewFactory(appLogger, nil)

	appLogger.Info("Creating WhatsApp manager with database connection...")
	manager, err := factory.CreateManager(database.GetDB().DB)
	if err != nil {
		return nil, err
	}

	appLogger.Info("WhatsApp manager created successfully - whatsmeow tables are now available")
	return manager, nil
}

// showVersion displays version information
func showVersion() {
	fmt.Printf("zpwoot - WhatsApp Multi-Session API\n")
	fmt.Printf("Version: %s\n", Version)
	fmt.Printf("Build Time: %s\n", BuildTime)
	fmt.Printf("Git Commit: %s\n", GitCommit)
	fmt.Printf("Go Version: %s\n", runtime.Version())
	fmt.Printf("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
}

// showMigrationStatus displays the current migration status
func showMigrationStatus(migrations []*db.Migration, logger *logger.Logger) {
	fmt.Printf("Migration Status:\n")
	fmt.Printf("================\n\n")

	if len(migrations) == 0 {
		fmt.Printf("No migrations found.\n")
		return
	}

	for _, migration := range migrations {
		status := "PENDING"
		appliedAt := "Not applied"

		if migration.AppliedAt != nil {
			status = "APPLIED"
			appliedAt = migration.AppliedAt.Format("2006-01-02 15:04:05")
		}

		fmt.Printf("Version: %03d | Status: %-7s | Name: %s | Applied: %s\n",
			migration.Version, status, migration.Name, appliedAt)
	}
	fmt.Printf("\n")
}

// seedDatabase seeds the database with sample data
func seedDatabase(database *platformDB.DB, logger *logger.Logger) error {
	logger.Info("Starting database seeding...")

	// Sample session data
	sampleSessions := []map[string]interface{}{
		{
			"id":          "sample-session-1",
			"name":        "Sample WhatsApp Session",
			"device_jid":  "5511999999999@s.whatsapp.net",
			"status":      "created",
			"created_at":  time.Now(),
			"updated_at":  time.Now(),
		},
	}

	// Sample webhook data
	sampleWebhooks := []map[string]interface{}{
		{
			"id":          "sample-webhook-1",
			"session_id":  "sample-session-1",
			"url":         "https://example.com/webhook",
			"events":      []string{"message", "status"},
			"enabled":     true,
			"created_at":  time.Now(),
			"updated_at":  time.Now(),
		},
	}

	// Insert sample sessions
	for _, session := range sampleSessions {
		query := `
			INSERT INTO "zpSessions" ("id", "name", "deviceJid", "status", "createdAt", "updatedAt")
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT ("id") DO NOTHING
		`
		_, err := database.GetDB().Exec(query,
			session["id"], session["name"], session["device_jid"],
			session["status"], session["created_at"], session["updated_at"])
		if err != nil {
			return fmt.Errorf("failed to insert sample session: %w", err)
		}
	}

	// Insert sample webhooks
	for _, webhook := range sampleWebhooks {
		query := `
			INSERT INTO "zpWebhooks" ("id", "sessionId", "url", "events", "enabled", "createdAt", "updatedAt")
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT ("id") DO NOTHING
		`
		_, err := database.GetDB().Exec(query,
			webhook["id"], webhook["session_id"], webhook["url"],
			webhook["events"], webhook["enabled"], webhook["created_at"], webhook["updated_at"])
		if err != nil {
			return fmt.Errorf("failed to insert sample webhook: %w", err)
		}
	}

	logger.InfoWithFields("Database seeding completed", map[string]interface{}{
		"sessions_created": len(sampleSessions),
		"webhooks_created": len(sampleWebhooks),
	})

	return nil
}
