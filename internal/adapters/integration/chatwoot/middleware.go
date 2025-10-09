package chatwoot

import (
	"context"

	"zpwoot/internal/core/ports/output"
)

type WebhookMiddleware struct {
	integrator *Integrator
	logger     output.Logger
}

func NewWebhookMiddleware(integrator *Integrator, logger output.Logger) *WebhookMiddleware {
	return &WebhookMiddleware{
		integrator: integrator,
		logger:     logger,
	}
}

func (m *WebhookMiddleware) ProcessWebhook(ctx context.Context, sessionID string, event *output.WebhookEvent) error {

	m.logger.Debug().
		Str("session_id", sessionID).
		Str("event", event.Type).
		Msg("Processing webhook for Chatwoot integration")

	if err := m.integrator.ProcessWhatsAppEvent(ctx, sessionID, event); err != nil {
		m.logger.Error().
			Err(err).
			Str("session_id", sessionID).
			Str("event", event.Type).
			Msg("Failed to process webhook for Chatwoot")

	}

	return nil
}

func (m *WebhookMiddleware) ShouldProcess(event *output.WebhookEvent) bool {

	switch event.Type {
	case "message", "connection.update", "qr":
		return true
	default:
		return false
	}
}
