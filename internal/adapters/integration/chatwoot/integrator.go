package chatwoot

import (
	"context"

	"zpwoot/internal/core/domain/chatwoot"
	"zpwoot/internal/core/ports/output"
)

type Integrator struct {
	eventHandler   *EventHandler
	webhookHandler *WebhookHandler
	manager        *Manager
	logger         output.Logger
}

func NewIntegrator(
	chatwootRepo chatwoot.Repository,
	whatsappClient output.WhatsAppClient,
	logger output.Logger,
	baseURL string,
) *Integrator {
	manager := NewManager(chatwootRepo, whatsappClient, logger, baseURL)
	eventHandler := NewEventHandler(manager, logger)
	webhookHandler := NewWebhookHandler(manager, logger)

	return &Integrator{
		eventHandler:   eventHandler,
		webhookHandler: webhookHandler,
		manager:        manager,
		logger:         logger,
	}
}

func (i *Integrator) SetWhatsAppClient(client output.WhatsAppClient) {
	i.manager.SetWhatsAppClient(client)
}

func (i *Integrator) ProcessWhatsAppEvent(ctx context.Context, sessionID string, event *output.WebhookEvent) error {
	switch event.Type {
	case "message":
		return i.eventHandler.HandleMessage(ctx, sessionID, event)
	case "connection.update":
		return i.eventHandler.HandleConnectionUpdate(ctx, sessionID, event)
	case "qr":
		return i.eventHandler.HandleQRCode(ctx, sessionID, event)
	default:
		i.logger.Debug().
			Str("session_id", sessionID).
			Str("event", event.Type).
			Msg("Unhandled WhatsApp event for Chatwoot")
		return nil
	}
}

func (i *Integrator) ProcessChatwootWebhook(ctx context.Context, sessionID string, payload *WebhookPayload) (*WebhookResponse, error) {
	return i.webhookHandler.HandleWebhook(ctx, sessionID, payload)
}

func (i *Integrator) InitializeSession(ctx context.Context, sessionID string) error {
	return i.manager.InitializeInbox(ctx, sessionID)
}

func (i *Integrator) SyncContacts(ctx context.Context, sessionID string) error {
	return i.manager.SyncContacts(ctx, sessionID)
}

func (i *Integrator) GetManager() *Manager {
	return i.manager
}

func (i *Integrator) GetEventHandler() *EventHandler {
	return i.eventHandler
}

func (i *Integrator) GetWebhookHandler() *WebhookHandler {
	return i.webhookHandler
}
