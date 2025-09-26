// Package main provides the entry point for ZPMeow application
//
// @title ZPMeow - WhatsApp Multi-Session API
// @version 1.0
// @description A complete REST API for managing multiple WhatsApp sessions using Go, Fiber, PostgreSQL, and whatsmeow library.
// @description
// @description ## Authentication
// @description All API endpoints (except /health and /swagger/*) require API key authentication.
// @description Provide your API key in the `Authorization` header.
//
// @contact.name ZPMeow Support
// @contact.url https://github.com/your-org/zpmeow
// @contact.email support@zpmeow.com
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
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"

	_ "zpwoot/docs/swagger" // Import generated swagger docs
	"zpwoot/internal/app"
	"zpwoot/internal/infra/http/middleware"
	"zpwoot/internal/infra/http/routers"
	"zpwoot/internal/infra/repository"
	"zpwoot/internal/infra/wmeow"
	"zpwoot/platform/config"
	"zpwoot/platform/db"
	"zpwoot/platform/logger"
)

func main() {
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
	database, err := db.NewWithMigrations(cfg.DatabaseURL, appLogger)
	if err != nil {
		appLogger.Fatal("Failed to connect to database and run migrations: " + err.Error())
	}
	defer func() {
		if err := database.Close(); err != nil {
			appLogger.Error("Failed to close database connection: " + err.Error())
		}
	}()

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
		ChatwootIntegration: nil, // TODO: implement when needed
		Logger:              appLogger,
		Version:             "1.0.0", // TODO: get from build flags
		BuildTime:           "dev",   // TODO: get from build flags
		GitCommit:           "dev",   // TODO: get from build flags
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
	appLogger.InfoWithFields("Starting ZPMeow server", map[string]interface{}{
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
