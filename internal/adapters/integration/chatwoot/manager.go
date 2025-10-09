package chatwoot

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"zpwoot/internal/core/domain/chatwoot"
	"zpwoot/internal/core/ports/output"
)

type Manager struct {
	service        *Service
	chatwootRepo   chatwoot.Repository
	whatsappClient output.WhatsAppClient
	logger         output.Logger
	baseURL        string
}

func NewManager(
	chatwootRepo chatwoot.Repository,
	whatsappClient output.WhatsAppClient,
	logger output.Logger,
	baseURL string,
) *Manager {
	return &Manager{
		service:        NewService(logger),
		chatwootRepo:   chatwootRepo,
		whatsappClient: whatsappClient,
		logger:         logger,
		baseURL:        baseURL,
	}
}

func (m *Manager) InitializeInbox(ctx context.Context, sessionID string) error {

	config, err := m.chatwootRepo.GetBySessionID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to get Chatwoot config: %w", err)
	}

	if !config.Enabled {
		m.logger.Debug().
			Str("session_id", sessionID).
			Msg("Chatwoot integration disabled for session")
		return nil
	}

	phoneNumber := sessionID

	inbox, err := m.service.SetupInbox(ctx, config, sessionID, phoneNumber, m.baseURL)
	if err != nil {
		return fmt.Errorf("failed to setup inbox: %w", err)
	}

	inboxIDStr := strconv.Itoa(inbox.ID)
	if config.InboxID == nil || *config.InboxID != inboxIDStr {
		config.InboxID = &inboxIDStr
		config.InboxName = &inbox.Name
		config.UpdatedAt = time.Now()

		if err := m.chatwootRepo.Update(ctx, config); err != nil {
			m.logger.Error().
				Err(err).
				Str("session_id", sessionID).
				Int("inbox_id", inbox.ID).
				Msg("Failed to update Chatwoot config with inbox ID")
		}
	}

	m.logger.Info().
		Str("session_id", sessionID).
		Int("inbox_id", inbox.ID).
		Str("phone_number", phoneNumber).
		Msg("Chatwoot inbox initialized successfully")

	return nil
}

func (m *Manager) ProcessIncomingMessage(ctx context.Context, sessionID string, message *WhatsAppMessage) error {

	config, err := m.chatwootRepo.GetBySessionID(ctx, sessionID)
	if err != nil {
		m.logger.Debug().
			Err(err).
			Str("session_id", sessionID).
			Msg("No Chatwoot config found for session")
		return nil
	}

	if !config.Enabled {
		m.logger.Debug().
			Str("session_id", sessionID).
			Msg("Chatwoot integration disabled for session")
		return nil
	}

	if m.shouldIgnoreMessage(config, message) {
		m.logger.Debug().
			Str("session_id", sessionID).
			Str("from", message.From).
			Msg("Message ignored based on configuration")
		return nil
	}

	params := SendMessageParams{
		SessionID:   sessionID,
		MessageID:   message.ID,
		ChatID:      message.ChatID,
		From:        message.From,
		FromName:    message.FromName,
		Content:     message.Content,
		MessageType: message.Type,
		MediaURL:    message.MediaURL,
		MediaType:   message.MediaType,
		FileName:    message.FileName,
		Timestamp:   message.Timestamp,
	}

	if err := m.service.SendMessageToChatwoot(ctx, config, params); err != nil {
		m.logger.Error().
			Err(err).
			Str("session_id", sessionID).
			Str("message_id", message.ID).
			Msg("Failed to send message to Chatwoot")
		return fmt.Errorf("failed to send message to Chatwoot: %w", err)
	}

	return nil
}

func (m *Manager) ProcessChatwootWebhook(ctx context.Context, sessionID string, payload *WebhookPayload) (*WebhookResponse, error) {

	config, err := m.chatwootRepo.GetBySessionID(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get Chatwoot config: %w", err)
	}

	if !config.Enabled {
		return &WebhookResponse{
			Success: false,
			Message: "Chatwoot integration disabled for session",
		}, nil
	}

	response, err := m.service.ProcessChatwootWebhook(ctx, config, payload)
	if err != nil {
		m.logger.Error().
			Err(err).
			Str("session_id", sessionID).
			Str("event", payload.Event).
			Msg("Failed to process Chatwoot webhook")
		return nil, fmt.Errorf("failed to process webhook: %w", err)
	}

	if payload.Event == "message_created" && response.Success {
		if err := m.handleOutgoingMessage(ctx, sessionID, response.Data); err != nil {
			m.logger.Error().
				Err(err).
				Str("session_id", sessionID).
				Msg("Failed to send outgoing message via WhatsApp")
		}
	}

	return response, nil
}

func (m *Manager) handleOutgoingMessage(ctx context.Context, sessionID string, data map[string]interface{}) error {
	if data == nil {
		return nil
	}

	whatsappNumber, ok := data["whatsapp_number"].(string)
	if !ok || whatsappNumber == "" {
		return fmt.Errorf("missing whatsapp_number in response data")
	}

	chatID, ok := data["chat_id"].(string)
	if !ok || chatID == "" {
		return fmt.Errorf("missing chat_id in response data")
	}

	content, ok := data["content"].(string)
	if !ok || content == "" {
		return fmt.Errorf("missing content in response data")
	}

	_, err := m.whatsappClient.SendTextMessage(ctx, sessionID, chatID, content)
	if err != nil {
		return fmt.Errorf("failed to send WhatsApp message: %w", err)
	}

	m.logger.Info().
		Str("session_id", sessionID).
		Str("to", chatID).
		Str("content", content).
		Msg("Message sent from Chatwoot to WhatsApp")

	return nil
}

func (m *Manager) shouldIgnoreMessage(config *chatwoot.Chatwoot, message *WhatsAppMessage) bool {

	for _, ignoreJID := range config.IgnoreJids {
		if strings.Contains(message.From, ignoreJID) || strings.Contains(message.ChatID, ignoreJID) {
			return true
		}
	}

	if strings.Contains(message.ChatID, "status@broadcast") {
		return true
	}

	if message.IsGroup && !config.ImportMessages {
		return true
	}

	return false
}

func (m *Manager) SyncContacts(ctx context.Context, sessionID string) error {

	config, err := m.chatwootRepo.GetBySessionID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to get Chatwoot config: %w", err)
	}

	if !config.Enabled || !config.ImportContacts {
		m.logger.Debug().
			Str("session_id", sessionID).
			Msg("Contact sync disabled for session")
		return nil
	}

	m.logger.Info().
		Str("session_id", sessionID).
		Msg("Contact synchronization not yet implemented")

	return nil
}

type WhatsAppMessage struct {
	ID        string
	ChatID    string
	From      string
	FromName  string
	Content   string
	Type      string
	MediaURL  string
	MediaType string
	FileName  string
	IsGroup   bool
	Timestamp int64
}

type Contact struct {
	PhoneNumber string
	Name        string
}
