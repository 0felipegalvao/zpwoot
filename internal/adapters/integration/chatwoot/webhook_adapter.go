package chatwoot

import (
	"context"

	"zpwoot/internal/core/application/dto"
	"zpwoot/internal/core/ports/input"
	"zpwoot/internal/core/ports/output"
)

type WebhookAdapter struct {
	handler *WebhookHandler
	logger  output.Logger
}

func NewWebhookAdapter(handler *WebhookHandler, logger output.Logger) input.ChatwootWebhookHandler {
	return &WebhookAdapter{
		handler: handler,
		logger:  logger,
	}
}

func (a *WebhookAdapter) HandleWebhook(ctx context.Context, sessionID string, req *dto.ChatwootWebhookRequest) (*dto.ChatwootWebhookResponse, error) {

	payload := &WebhookPayload{
		Event:     req.Event,
		Timestamp: 0,
		Data: map[string]interface{}{
			"account":      req.Account,
			"conversation": req.Conversation,
			"message":      req.Message,
			"contact":      req.Contact,
		},
	}

	response, err := a.handler.HandleWebhook(ctx, sessionID, payload)
	if err != nil {
		return &dto.ChatwootWebhookResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &dto.ChatwootWebhookResponse{
		Success: response.Success,
		Message: response.Message,
	}, nil
}
