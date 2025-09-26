package app

import (
	"zpwoot/internal/domain/chatwoot"
	"zpwoot/internal/domain/session"
	"zpwoot/internal/domain/webhook"
	"zpwoot/internal/ports"
	"zpwoot/platform/logger"
)

// Container holds all use cases and their dependencies
type Container struct {
	// Use Cases
	CommonUseCase   CommonUseCase
	SessionUseCase  SessionUseCase
	WebhookUseCase  WebhookUseCase
	ChatwootUseCase ChatwootUseCase

	// Dependencies
	logger      *logger.Logger
	sessionRepo ports.SessionRepository
}

// ContainerConfig holds configuration for creating the container
type ContainerConfig struct {
	// Repositories
	SessionRepo  ports.SessionRepository
	WebhookRepo  ports.WebhookRepository
	ChatwootRepo ports.ChatwootRepository

	// External integrations
	WhatsAppManager     ports.WhatsAppManager
	ChatwootIntegration ports.ChatwootIntegration

	// Infrastructure
	Logger *logger.Logger

	// Application metadata
	Version   string
	BuildTime string
	GitCommit string
}

// NewContainer creates a new application container with all use cases
func NewContainer(config *ContainerConfig) *Container {
	// Create domain services
	sessionService := session.NewService(
		config.SessionRepo,
		config.WhatsAppManager,
	)

	webhookService := webhook.NewService(
		config.Logger,
	)

	chatwootService := chatwoot.NewService(
		config.Logger,
	)

	// Create use cases
	commonUseCase := NewCommonUseCase(
		config.Version,
		config.BuildTime,
		config.GitCommit,
	)

	sessionUseCase := NewSessionUseCase(
		config.SessionRepo,
		config.WhatsAppManager,
		sessionService,
	)

	webhookUseCase := NewWebhookUseCase(
		config.WebhookRepo,
		webhookService,
	)

	chatwootUseCase := NewChatwootUseCase(
		config.ChatwootRepo,
		config.ChatwootIntegration,
		chatwootService,
	)

	return &Container{
		CommonUseCase:   commonUseCase,
		SessionUseCase:  sessionUseCase,
		WebhookUseCase:  webhookUseCase,
		ChatwootUseCase: chatwootUseCase,
		logger:          config.Logger,
		sessionRepo:     config.SessionRepo,
	}
}

// GetCommonUseCase returns the common use case
func (c *Container) GetCommonUseCase() CommonUseCase {
	return c.CommonUseCase
}

// GetSessionUseCase returns the session use case
func (c *Container) GetSessionUseCase() SessionUseCase {
	return c.SessionUseCase
}

// GetWebhookUseCase returns the webhook use case
func (c *Container) GetWebhookUseCase() WebhookUseCase {
	return c.WebhookUseCase
}

// GetChatwootUseCase returns the chatwoot use case
func (c *Container) GetChatwootUseCase() ChatwootUseCase {
	return c.ChatwootUseCase
}

// GetLogger returns the logger instance
func (c *Container) GetLogger() *logger.Logger {
	return c.logger
}

// GetSessionRepository returns the session repository instance
func (c *Container) GetSessionRepository() ports.SessionRepository {
	return c.sessionRepo
}
