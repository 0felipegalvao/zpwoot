package chatwoot

import (
	"context"
	"fmt"
	"strings"
	"time"

	"zpwoot/internal/core/ports/output"
)

type EventHandler struct {
	manager *Manager
	logger  output.Logger
}

func NewEventHandler(manager *Manager, logger output.Logger) *EventHandler {
	return &EventHandler{
		manager: manager,
		logger:  logger,
	}
}

func (h *EventHandler) HandleMessage(ctx context.Context, sessionID string, event *output.WebhookEvent) error {

	messageData, ok := event.Data["message"].(map[string]interface{})
	if !ok {
		h.logger.Debug().
			Str("session_id", sessionID).
			Msg("No message data in webhook event")
		return nil
	}

	chatID, _ := messageData["chat_id"].(string)
	from, _ := messageData["from"].(string)
	fromName, _ := messageData["from_name"].(string)
	messageID, _ := messageData["id"].(string)
	isGroup, _ := messageData["is_group"].(bool)

	if strings.Contains(chatID, "status@broadcast") {
		return nil
	}

	content := h.extractMessageContent(messageData)
	messageType := h.getMessageType(messageData)
	mediaURL := h.getMediaURL(messageData)
	mediaType := h.getMediaType(messageData)
	fileName := h.getFileName(messageData)

	message := &WhatsAppMessage{
		ID:        messageID,
		ChatID:    chatID,
		From:      from,
		FromName:  fromName,
		Content:   content,
		Type:      messageType,
		MediaURL:  mediaURL,
		MediaType: mediaType,
		FileName:  fileName,
		IsGroup:   isGroup,
		Timestamp: time.Now().Unix(),
	}

	return h.manager.ProcessIncomingMessage(ctx, sessionID, message)
}

func (h *EventHandler) HandleConnectionUpdate(ctx context.Context, sessionID string, event *output.WebhookEvent) error {
	statusData, ok := event.Data["status"].(map[string]interface{})
	if !ok {
		return nil
	}

	status, _ := statusData["state"].(string)

	h.logger.Info().
		Str("session_id", sessionID).
		Str("status", status).
		Msg("WhatsApp connection status changed")

	if status == "open" {
		if err := h.manager.InitializeInbox(ctx, sessionID); err != nil {
			h.logger.Error().
				Err(err).
				Str("session_id", sessionID).
				Msg("Failed to initialize Chatwoot inbox")
			return err
		}
	}

	return nil
}

func (h *EventHandler) HandleQRCode(ctx context.Context, sessionID string, event *output.WebhookEvent) error {
	qrData, ok := event.Data["qr"].(map[string]interface{})
	if !ok {
		return nil
	}

	status, _ := qrData["status"].(string)

	h.logger.Info().
		Str("session_id", sessionID).
		Str("status", status).
		Msg("QR code event received")

	return nil
}

func (h *EventHandler) extractMessageContent(messageData map[string]interface{}) string {

	if content, ok := messageData["content"].(string); ok && content != "" {
		return content
	}

	if text, ok := messageData["text"].(string); ok && text != "" {
		return text
	}

	if body, ok := messageData["body"].(string); ok && body != "" {
		return body
	}

	if caption, ok := messageData["caption"].(string); ok && caption != "" {
		return caption
	}

	if location, ok := messageData["location"].(map[string]interface{}); ok {
		return h.formatLocationMessage(location)
	}

	if contact, ok := messageData["contact"].(map[string]interface{}); ok {
		return h.formatContactMessage(contact)
	}

	if messageType, ok := messageData["type"].(string); ok {
		switch messageType {
		case "image", "video", "audio", "document":
			return "[Media]"
		case "sticker":
			return "[Sticker]"
		case "location":
			return "[Location]"
		case "contact":
			return "[Contact]"
		}
	}

	return ""
}

func (h *EventHandler) getMessageType(messageData map[string]interface{}) string {
	if msgType, ok := messageData["type"].(string); ok {
		return msgType
	}
	return "text"
}

func (h *EventHandler) getMediaURL(messageData map[string]interface{}) string {
	if media, ok := messageData["media"].(map[string]interface{}); ok {
		if url, ok := media["url"].(string); ok {
			return url
		}
	}
	return ""
}

func (h *EventHandler) getMediaType(messageData map[string]interface{}) string {
	if media, ok := messageData["media"].(map[string]interface{}); ok {
		if mimeType, ok := media["mime_type"].(string); ok {
			return mimeType
		}
	}
	return ""
}

func (h *EventHandler) getFileName(messageData map[string]interface{}) string {
	if media, ok := messageData["media"].(map[string]interface{}); ok {
		if fileName, ok := media["filename"].(string); ok {
			return fileName
		}
	}
	return ""
}

func (h *EventHandler) formatLocationMessage(location map[string]interface{}) string {
	lat, _ := location["latitude"].(string)
	lng, _ := location["longitude"].(string)
	name, _ := location["name"].(string)
	address, _ := location["address"].(string)

	content := "📍 **Location:**\n\n"
	content += fmt.Sprintf("**Latitude:** %s\n", lat)
	content += fmt.Sprintf("**Longitude:** %s\n", lng)

	if name != "" {
		content += fmt.Sprintf("**Name:** %s\n", name)
	}

	if address != "" {
		content += fmt.Sprintf("**Address:** %s\n", address)
	}

	content += fmt.Sprintf("**URL:** https://www.google.com/maps/search/?api=1&query=%s,%s", lat, lng)

	return content
}

func (h *EventHandler) formatContactMessage(contact map[string]interface{}) string {
	name, _ := contact["name"].(string)
	phone, _ := contact["phone"].(string)
	email, _ := contact["email"].(string)

	content := "👤 **Contact:**\n\n"
	content += fmt.Sprintf("**Name:** %s\n", name)

	if phone != "" {
		content += fmt.Sprintf("**Phone:** %s\n", phone)
	}

	if email != "" {
		content += fmt.Sprintf("**Email:** %s\n", email)
	}

	return content
}

type WebhookHandler struct {
	manager *Manager
	logger  output.Logger
}

func NewWebhookHandler(manager *Manager, logger output.Logger) *WebhookHandler {
	return &WebhookHandler{
		manager: manager,
		logger:  logger,
	}
}

func (h *WebhookHandler) HandleWebhook(ctx context.Context, sessionID string, payload *WebhookPayload) (*WebhookResponse, error) {
	h.logger.Info().
		Str("session_id", sessionID).
		Str("event", payload.Event).
		Int64("timestamp", payload.Timestamp).
		Msg("Processing Chatwoot webhook")

	return h.manager.ProcessChatwootWebhook(ctx, sessionID, payload)
}

func (h *WebhookHandler) ValidateWebhook(payload *WebhookPayload) error {
	if payload.Event == "" {
		return fmt.Errorf("missing event in webhook payload")
	}

	if payload.Data == nil {
		return fmt.Errorf("missing data in webhook payload")
	}

	return nil
}

func (h *WebhookHandler) GetSupportedEvents() []string {
	return []string{
		"message_created",
		"conversation_status_changed",
		"conversation_updated",
	}
}

func (h *WebhookHandler) IsEventSupported(event string) bool {
	supportedEvents := h.GetSupportedEvents()
	for _, supportedEvent := range supportedEvents {
		if supportedEvent == event {
			return true
		}
	}
	return false
}
