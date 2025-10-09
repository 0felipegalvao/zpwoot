package chatwoot

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"zpwoot/internal/core/domain/chatwoot"
	"zpwoot/internal/core/ports/output"
)

type Service struct {
	logger output.Logger
}

func NewService(logger output.Logger) *Service {
	return &Service{
		logger: logger,
	}
}

func (s *Service) CreateClient(config *chatwoot.Chatwoot) *Client {
	return NewClient(config.URL, config.Token, config.AccountID, s.logger)
}

func (s *Service) SetupInbox(ctx context.Context, config *chatwoot.Chatwoot, sessionID, phoneNumber, baseURL string) (*InboxResponse, error) {
	client := s.CreateClient(config)

	if config.InboxID != nil && *config.InboxID != "" {
		inboxID, err := strconv.Atoi(*config.InboxID)
		if err == nil {
			inbox, err := client.GetInbox(ctx, inboxID)
			if err == nil {
				s.logger.Info().
					Str("session_id", sessionID).
					Int("inbox_id", inboxID).
					Msg("Using existing Chatwoot inbox")
				return inbox, nil
			}
		}
	}

	webhookURL := s.BuildWebhookURL(baseURL, sessionID)

	inboxName := fmt.Sprintf("WhatsApp - %s", phoneNumber)
	if config.InboxName != nil && *config.InboxName != "" {
		inboxName = *config.InboxName
	}

	inboxReq := &InboxRequest{
		Name:        inboxName,
		Channel:     "api",
		PhoneNumber: phoneNumber,
		Provider:    "whatsapp",
		WebhookURL:  webhookURL,
	}

	inbox, err := client.CreateInbox(ctx, inboxReq)
	if err != nil {
		s.logger.Error().
			Err(err).
			Str("session_id", sessionID).
			Msg("Failed to create Chatwoot inbox")
		return nil, fmt.Errorf("failed to create inbox: %w", err)
	}

	s.logger.Info().
		Str("session_id", sessionID).
		Int("inbox_id", inbox.ID).
		Str("inbox_name", inbox.Name).
		Str("webhook_url", webhookURL).
		Msg("Created new Chatwoot inbox")

	return inbox, nil
}

func (s *Service) CreateOrUpdateContact(ctx context.Context, client *Client, phoneNumber, name string) (*ContactResponse, error) {
	// Extract and format phone number from WhatsApp JID
	extractedPhone := s.extractPhoneNumberFromJID(phoneNumber)
	formattedPhone := s.formatPhoneNumberForChatwoot(extractedPhone)

	// First try to find contact by formatted phone number
	s.logger.Debug().
		Str("phone_number", formattedPhone).
		Msg("Searching for existing Chatwoot contact")

	contact, err := client.GetContactByIdentifier(ctx, formattedPhone)
	if err == nil {
		s.logger.Debug().
			Str("phone_number", formattedPhone).
			Int("contact_id", contact.ID).
			Msg("Found existing Chatwoot contact")
		return contact, nil
	}

	s.logger.Debug().
		Str("phone_number", formattedPhone).
		Err(err).
		Msg("Contact not found, will create new one")

	contactReq := &ContactRequest{
		Name:        name,
		PhoneNumber: formattedPhone,
		Identifier:  phoneNumber, // Use original JID as identifier
		CustomAttributes: map[string]interface{}{
			"whatsapp_number": phoneNumber, // Keep original JID for reference
		},
	}

	contact, err = client.CreateContact(ctx, contactReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create contact: %w", err)
	}

	s.logger.Info().
		Str("phone_number", phoneNumber).
		Int("contact_id", contact.ID).
		Str("contact_name", contact.Name).
		Msg("Created new Chatwoot contact")

	return contact, nil
}

func (s *Service) CreateOrGetConversation(ctx context.Context, client *Client, sourceID string, inboxID, contactID int) (*ConversationResponse, error) {

	convReq := &ConversationRequest{
		SourceID:  sourceID,
		InboxID:   inboxID,
		ContactID: contactID,
		Status:    "open",
	}

	conversation, err := client.CreateConversation(ctx, convReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create/get conversation: %w", err)
	}

	s.logger.Debug().
		Str("source_id", sourceID).
		Int("conversation_id", conversation.ID).
		Int("inbox_id", inboxID).
		Int("contact_id", contactID).
		Msg("Created/retrieved Chatwoot conversation")

	return conversation, nil
}

func (s *Service) SendMessageToChatwoot(ctx context.Context, config *chatwoot.Chatwoot, params SendMessageParams) error {
	client := s.CreateClient(config)

	inboxID, err := strconv.Atoi(*config.InboxID)
	if err != nil {
		return fmt.Errorf("invalid inbox ID: %w", err)
	}

	contact, err := s.CreateOrUpdateContact(ctx, client, params.From, params.FromName)
	if err != nil {
		return fmt.Errorf("failed to handle contact: %w", err)
	}

	sourceID := fmt.Sprintf("%s_%s", params.From, params.ChatID)
	conversation, err := s.CreateOrGetConversation(ctx, client, sourceID, inboxID, contact.ID)
	if err != nil {
		return fmt.Errorf("failed to handle conversation: %w", err)
	}

	content := params.Content
	if params.MessageType == "media" && params.MediaURL != "" {
		content = fmt.Sprintf("%s\n\n📎 %s", content, params.MediaURL)
	}

	msgReq := &MessageRequest{
		Content:     content,
		MessageType: "incoming",
		ContentType: "text",
		Echo: map[string]interface{}{
			"whatsapp_message_id": params.MessageID,
			"whatsapp_chat_id":    params.ChatID,
			"message_type":        params.MessageType,
		},
	}

	if params.MediaURL != "" {
		msgReq.Attachments = []AttachmentRequest{
			{
				Content:     params.MediaURL,
				ContentType: params.MediaType,
				FileName:    params.FileName,
			},
		}
	}

	message, err := client.SendMessage(ctx, conversation.ID, msgReq)
	if err != nil {
		return fmt.Errorf("failed to send message to Chatwoot: %w", err)
	}

	s.logger.Info().
		Str("session_id", params.SessionID).
		Str("from", params.From).
		Str("chat_id", params.ChatID).
		Int("conversation_id", conversation.ID).
		Int("message_id", message.ID).
		Msg("Message sent to Chatwoot successfully")

	return nil
}

type SendMessageParams struct {
	SessionID   string
	MessageID   string
	ChatID      string
	From        string
	FromName    string
	Content     string
	MessageType string
	MediaURL    string
	MediaType   string
	FileName    string
	Timestamp   int64
}

func (s *Service) ProcessChatwootWebhook(ctx context.Context, config *chatwoot.Chatwoot, webhook *WebhookPayload) (*WebhookResponse, error) {
	s.logger.Info().
		Str("event", webhook.Event).
		Interface("data", webhook.Data).
		Msg("Processing Chatwoot webhook")

	switch webhook.Event {
	case "message_created":
		return s.handleMessageCreated(ctx, config, webhook)
	case "conversation_status_changed":
		return s.handleConversationStatusChanged(ctx, config, webhook)
	default:
		s.logger.Debug().
			Str("event", webhook.Event).
			Msg("Unhandled Chatwoot webhook event")
		return &WebhookResponse{
			Success: true,
			Message: "Event received but not processed",
		}, nil
	}
}

func (s *Service) handleMessageCreated(ctx context.Context, config *chatwoot.Chatwoot, webhook *WebhookPayload) (*WebhookResponse, error) {
	_ = ctx
	_ = config

	messageData, ok := webhook.Data["message"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid message data in webhook")
	}

	messageType, _ := messageData["message_type"].(string)
	if messageType != "outgoing" {
		return &WebhookResponse{
			Success: true,
			Message: "Incoming message ignored",
		}, nil
	}

	conversationData, ok := webhook.Data["conversation"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid conversation data in webhook")
	}

	sourceID, _ := conversationData["source_id"].(string)
	if sourceID == "" {
		return nil, fmt.Errorf("missing source_id in conversation")
	}

	parts := strings.Split(sourceID, "_")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid source_id format: %s", sourceID)
	}

	whatsappNumber := parts[0]
	chatID := strings.Join(parts[1:], "_")

	content, _ := messageData["content"].(string)
	if content == "" {
		return &WebhookResponse{
			Success: true,
			Message: "Empty message content",
		}, nil
	}

	s.logger.Info().
		Str("whatsapp_number", whatsappNumber).
		Str("chat_id", chatID).
		Str("content", content).
		Msg("Processing outgoing message from Chatwoot")

	return &WebhookResponse{
		Success: true,
		Message: "Message processed successfully",
		Data: map[string]interface{}{
			"whatsapp_number": whatsappNumber,
			"chat_id":         chatID,
			"content":         content,
		},
	}, nil
}

func (s *Service) handleConversationStatusChanged(ctx context.Context, config *chatwoot.Chatwoot, webhook *WebhookPayload) (*WebhookResponse, error) {
	_ = ctx
	_ = config

	conversationData, ok := webhook.Data["conversation"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid conversation data in webhook")
	}

	status, _ := conversationData["status"].(string)
	conversationID, _ := conversationData["id"].(float64)

	s.logger.Info().
		Float64("conversation_id", conversationID).
		Str("status", status).
		Msg("Conversation status changed")

	return &WebhookResponse{
		Success: true,
		Message: "Status change processed",
		Data: map[string]interface{}{
			"conversation_id": conversationID,
			"status":          status,
		},
	}, nil
}

type WebhookPayload struct {
	Event     string                 `json:"event"`
	Data      map[string]interface{} `json:"data"`
	Timestamp int64                  `json:"timestamp"`
}

type WebhookResponse struct {
	Success bool                   `json:"success"`
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

// BuildWebhookURL builds the webhook URL for a session
// Format: {baseURL}/sessions/{sessionID}/chatwoot/webhook
func (s *Service) BuildWebhookURL(baseURL, sessionID string) string {
	// Remove trailing slash from baseURL if present
	baseURL = strings.TrimSuffix(baseURL, "/")

	return fmt.Sprintf("%s/sessions/%s/chatwoot/webhook", baseURL, sessionID)
}

func (s *Service) FormatWhatsAppMessage(content, fromName, phoneNumber string, isGroup bool) string {
	if !isGroup {
		return content
	}

	formattedPhone := s.formatPhoneNumber(phoneNumber)
	return fmt.Sprintf("**%s - %s:**\n\n%s", formattedPhone, fromName, content)
}

func (s *Service) formatPhoneNumber(phoneNumber string) string {

	cleaned := strings.ReplaceAll(phoneNumber, "+", "")

	if strings.HasPrefix(cleaned, "55") && len(cleaned) >= 12 {

		if len(cleaned) == 13 {
			return fmt.Sprintf("+%s (%s) %s-%s",
				cleaned[:2], cleaned[2:4], cleaned[4:9], cleaned[9:])
		}

		if len(cleaned) == 12 {
			return fmt.Sprintf("+%s (%s) %s-%s",
				cleaned[:2], cleaned[2:4], cleaned[4:8], cleaned[8:])
		}
	}

	return "+" + cleaned
}

func (s *Service) GetMessageType(isFromMe bool) string {
	if isFromMe {
		return "outgoing"
	}
	return "incoming"
}

// extractPhoneNumberFromJID extracts phone number from WhatsApp JID
// Example: "559981769536:83@s.whatsapp.net" -> "559981769536"
func (s *Service) extractPhoneNumberFromJID(jid string) string {
	// Remove device ID part (e.g., ":83")
	withoutDevice := strings.Split(jid, ":")[0]
	// Remove domain part (e.g., "@s.whatsapp.net")
	phoneNumber := strings.Split(withoutDevice, "@")[0]
	return phoneNumber
}

// formatPhoneNumberForChatwoot formats phone number for Chatwoot (E.164 format)
func (s *Service) formatPhoneNumberForChatwoot(phoneNumber string) string {
	// Remove any existing + sign
	cleaned := strings.TrimPrefix(phoneNumber, "+")
	// Add + prefix for E.164 format
	return "+" + cleaned
}
