package routers

import (
	"github.com/gofiber/fiber/v2"
	fiberSwagger "github.com/swaggo/fiber-swagger"

	"zpwoot/internal/app"
	"zpwoot/internal/infra/http/handlers"
	"zpwoot/internal/infra/wmeow"
	"zpwoot/platform/db"
	"zpwoot/platform/logger"
)

// SetupRoutes configures all application routes
func SetupRoutes(app *fiber.App, database *db.DB, logger *logger.Logger, whatsappManager *wmeow.Manager, container *app.Container) {
	// Swagger documentation
	app.Get("/swagger/*", fiberSwagger.WrapHandler)

	// Health check
	// @Summary Health check
	// @Description Check if the API is running and healthy
	// @Tags Health
	// @Produce json
	// @Success 200 {object} object "API is healthy"
	// @Router /health [get]
	app.Get("/health", func(c *fiber.Ctx) error {
		response := &app.HealthResponse{
			Status:  "ok",
			Service: "zpwoot",
		}
		return c.JSON(response)
	})

	// WhatsApp health check
	// @Summary WhatsApp health check
	// @Description Check if WhatsApp manager and whatsmeow tables are available
	// @Tags Health
	// @Produce json
	// @Success 200 {object} object "WhatsApp manager is healthy"
	// @Router /health/whatsapp [get]
	app.Get("/health/whatsapp", func(c *fiber.Ctx) error {
		if whatsappManager == nil {
			return c.Status(503).JSON(fiber.Map{
				"status":  "error",
				"service": "whatsapp",
				"message": "WhatsApp manager not initialized",
			})
		}

		// Get health check from manager
		healthData := whatsappManager.HealthCheck()
		healthData["service"] = "whatsapp"
		healthData["message"] = "WhatsApp manager is healthy and whatsmeow tables are available"

		return c.JSON(healthData)
	})

	// Session management routes
	setupSessionRoutes(app, database, logger, whatsappManager, container)

	// Session-specific routes (grouped by session ID)
	setupSessionSpecificRoutes(app, database, logger, whatsappManager, container)

	// Global webhook and chatwoot configuration routes
	setupGlobalRoutes(app, database, logger, whatsappManager, container)
}

// setupSessionRoutes configures session management routes
func setupSessionRoutes(app *fiber.App, database *db.DB, appLogger *logger.Logger, whatsappManager *wmeow.Manager, container *app.Container) {
	// Initialize session handler with use case and repository from container
	sessionHandler := handlers.NewSessionHandler(appLogger, container.GetSessionUseCase(), container.GetSessionRepository())

	// Log WhatsApp manager availability
	if whatsappManager != nil {
		appLogger.Info("WhatsApp manager is available for session routes")
	} else {
		appLogger.Warn("WhatsApp manager is nil - session functionality will be limited")
	}

	sessions := app.Group("/sessions")

	// Session management routes (supports both UUID and session names)
	sessions.Post("/create", sessionHandler.CreateSession)              // POST /sessions/create
	sessions.Get("/list", sessionHandler.ListSessions)                  // GET /sessions/list
	sessions.Get("/:sessionId/info", sessionHandler.GetSessionInfo)     // GET /sessions/:sessionId/info
	sessions.Delete("/:sessionId/delete", sessionHandler.DeleteSession) // DELETE /sessions/:sessionId/delete
	sessions.Post("/:sessionId/connect", sessionHandler.ConnectSession) // POST /sessions/:sessionId/connect
	sessions.Post("/:sessionId/logout", sessionHandler.LogoutSession)   // POST /sessions/:sessionId/logout
	sessions.Get("/:sessionId/qr", sessionHandler.GetQRCode)            // GET /sessions/:sessionId/qr
	sessions.Post("/:sessionId/pair", sessionHandler.PairPhone)         // POST /sessions/:sessionId/pair
	sessions.Post("/:sessionId/proxy/set", sessionHandler.SetProxy)     // POST /sessions/:sessionId/proxy/set
	sessions.Get("/:sessionId/proxy/find", sessionHandler.GetProxy)     // GET /sessions/:sessionId/proxy/find

	// Initialize webhook handler for session-specific routes
	webhookHandler := handlers.NewWebhookHandler(appLogger)

	// Session-specific webhook configuration (supports both UUID and session names)
	sessions.Post("/:sessionId/webhook/set", webhookHandler.SetConfig)  // POST /sessions/:sessionId/webhook/set
	sessions.Get("/:sessionId/webhook/find", webhookHandler.FindConfig) // GET /sessions/:sessionId/webhook/find

	// Session-specific Chatwoot configuration (simplified to 2 endpoints)
	chatwootHandler := handlers.NewChatwootHandler(container.GetChatwootUseCase(), appLogger)
	sessions.Post("/:sessionId/chatwoot/set", chatwootHandler.SetConfig)  // POST /sessions/:sessionId/chatwoot/set (create/update)
	sessions.Get("/:sessionId/chatwoot/find", chatwootHandler.FindConfig) // GET /sessions/:sessionId/chatwoot/find
}

// setupSessionSpecificRoutes configures routes grouped by session ID
func setupSessionSpecificRoutes(app *fiber.App, database *db.DB, appLogger *logger.Logger, whatsappManager *wmeow.Manager, container *app.Container) {
	// Placeholder for future session-specific routes if needed
	// Currently all required routes are in setupSessionRoutes
}

// setupGlobalRoutes configures global routes (currently none needed)
func setupGlobalRoutes(app *fiber.App, database *db.DB, appLogger *logger.Logger, whatsappManager *wmeow.Manager, container *app.Container) {
	// All configuration routes are now session-specific
	// This function is kept for future global routes if needed
}
