package waclient

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"zpwoot/internal/adapters/logger"
	"zpwoot/internal/core/domain/webhook"
	"zpwoot/internal/core/ports/output"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"go.mau.fi/whatsmeow/types/events"
)

const (
	unknownMessageType = "unknown"
)

type DefaultEventHandler struct {
	logger        *logger.Logger
	webhookSender output.WebhookSender
	webhookRepo   webhook.Repository
}

func NewDefaultEventHandler(logger *logger.Logger, webhookSender output.WebhookSender, webhookRepo webhook.Repository) *DefaultEventHandler {
	return &DefaultEventHandler{
		logger:        logger,
		webhookSender: webhookSender,
		webhookRepo:   webhookRepo,
	}
}

func (eh *DefaultEventHandler) HandleEvent(client *Client, event interface{}) error {
	switch evt := event.(type) {
	case *events.Message:
		return eh.handleMessage(client, evt)
	case *events.Receipt:
		return eh.handleReceipt(client, evt)
	case *events.Presence:
		return eh.handlePresence(client, evt)
	case *events.ChatPresence:
		return eh.handleChatPresence(client, evt)
	case *events.HistorySync:
		return eh.handleHistorySync(client, evt)
	case *events.AppStateSyncComplete:
		return eh.handleAppStateSyncComplete(client, evt)
	case *events.PushNameSetting:
		return eh.handlePushNameSetting(client, evt)
	case *events.BlocklistChange:
		return eh.handleBlocklistChange(client, evt)
	case *events.GroupInfo:
		return eh.handleGroupInfo(client, evt)
	case *events.JoinedGroup:
		return eh.handleJoinedGroup(client, evt)
	case *events.OfflineSyncPreview:
		return eh.handleOfflineSyncPreview(client, evt)
	default:

		if payload, err := json.Marshal(event); err == nil {
			log.Debug().
				Str("event", "unhandled").
				Str("type", fmt.Sprintf("%T", event)).
				Str("session_id", client.SessionID).
				RawJSON("payload", payload).
				Msg("Event received")
		}
		return nil
	}
}

func (eh *DefaultEventHandler) handleMessage(client *Client, evt *events.Message) error {

	if payload, err := json.Marshal(evt); err == nil {
		log.Info().
			Str("chat", evt.Info.Chat.String()).
			Str("from", evt.Info.Sender.String()).
			Str("type", getMessageType(evt.Message)).
			Bool("from_me", evt.Info.IsFromMe).
			Bool("is_group", evt.Info.IsGroup).
			Str("session_id", client.SessionID).
			RawJSON("payload", payload).
			Msg("Message received")
	}

	log.Debug().
		Str("pkg", "waclient").
		Str("session_id", client.SessionID).
		Msg("DEBUG: About to call sendWebhookIfEnabled for Message event")

	return eh.sendWebhookIfEnabled(client, EventMessage, evt)
}

func (eh *DefaultEventHandler) handleReceipt(client *Client, evt *events.Receipt) error {

	if payload, err := json.Marshal(evt); err == nil {
		log.Info().
			Str("chat", evt.Chat.String()).
			Str("from", evt.Sender.String()).
			Str("type", string(evt.Type)).
			Str("session_id", client.SessionID).
			RawJSON("payload", payload).
			Msg("Receipt received")
	}

	return eh.sendWebhookIfEnabled(client, EventReceipt, evt)
}

func (eh *DefaultEventHandler) handlePresence(client *Client, evt *events.Presence) error {

	if payload, err := json.Marshal(evt); err == nil {
		log.Debug().
			Str("event", "presence").
			Str("from", evt.From.String()).
			Str("session_id", client.SessionID).
			RawJSON("payload", payload).
			Msg("Event received")
	}

	return eh.sendWebhookIfEnabled(client, EventPresence, evt)
}

func (eh *DefaultEventHandler) handleChatPresence(client *Client, evt *events.ChatPresence) error {

	if payload, err := json.Marshal(evt); err == nil {
		log.Debug().
			Str("event", "chat_presence").
			Str("chat", evt.Chat.String()).
			Str("session_id", client.SessionID).
			RawJSON("payload", payload).
			Msg("Event received")
	}

	return eh.sendWebhookIfEnabled(client, EventChatPresence, evt)
}

func (eh *DefaultEventHandler) handleHistorySync(client *Client, evt *events.HistorySync) error {

	if payload, err := json.Marshal(evt); err == nil {
		log.Debug().
			Str("event", "history_sync").
			Str("session_id", client.SessionID).
			RawJSON("payload", payload).
			Msg("Event received")
	}

	return eh.sendWebhookIfEnabled(client, EventHistorySync, evt)
}

func (eh *DefaultEventHandler) handleAppStateSyncComplete(client *Client, evt *events.AppStateSyncComplete) error {
	eh.logger.Debug().
		Str("session_id", client.SessionID).
		Str("event", "app_state_sync").
		Msg("Event received")

	return nil
}

func (eh *DefaultEventHandler) handlePushNameSetting(client *Client, evt *events.PushNameSetting) error {
	eh.logger.Debug().
		Str("session_id", client.SessionID).
		Str("event", "push_name").
		Msg("Event received")

	return nil
}

func (eh *DefaultEventHandler) handleBlocklistChange(client *Client, evt *events.BlocklistChange) error {
	eh.logger.Debug().
		Str("session_id", client.SessionID).
		Str("event", "blocklist").
		Msg("Event received")

	return nil
}

func (eh *DefaultEventHandler) handleGroupInfo(client *Client, evt *events.GroupInfo) error {
	eh.logger.Debug().
		Str("session_id", client.SessionID).
		Str("event", "group_info").
		Msg("Event received")

	return nil
}

func (eh *DefaultEventHandler) handleJoinedGroup(client *Client, evt *events.JoinedGroup) error {
	eh.logger.Debug().
		Str("session_id", client.SessionID).
		Str("event", "joined_group").
		Msg("Event received")

	return nil
}

func (eh *DefaultEventHandler) handleOfflineSyncPreview(client *Client, evt *events.OfflineSyncPreview) error {
	eh.logger.Debug().
		Str("session_id", client.SessionID).
		Str("event", "offline_sync").
		Msg("Event received")

	return nil
}

func (eh *DefaultEventHandler) sendWebhookIfEnabled(client *Client, eventType EventType, eventData interface{}) error {
	log.Debug().
		Str("pkg", "waclient").
		Str("session_id", client.SessionID).
		Str("event_type", string(eventType)).
		Msg("DEBUG: sendWebhookIfEnabled called")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	webhookConfig, err := eh.webhookRepo.GetBySessionID(ctx, client.SessionID)
	if err != nil {
		if err.Error() == "webhook not found" {
			log.Debug().
				Str("pkg", "waclient").
				Str("session_id", client.SessionID).
				Msg("DEBUG: No webhook config found for session")
			return nil
		}
		eh.logger.Error().Err(err).Str("session_id", client.SessionID).Msg("Failed to load webhook config")
		return nil
	}

	log.Debug().
		Str("pkg", "waclient").
		Str("session_id", client.SessionID).
		Str("webhook_url", webhookConfig.URL).
		Bool("webhook_enabled", webhookConfig.Enabled).
		Interface("webhook_events", webhookConfig.Events).
		Msg("DEBUG: Webhook config loaded")

	if !eh.shouldSendWebhook(webhookConfig, eventType) {
		log.Debug().
			Str("pkg", "waclient").
			Str("session_id", client.SessionID).
			Str("event_type", string(eventType)).
			Msg("DEBUG: shouldSendWebhook returned false")
		return nil
	}

	log.Debug().
		Str("pkg", "waclient").
		Str("session_id", client.SessionID).
		Str("event_type", string(eventType)).
		Msg("DEBUG: About to send webhook")

	return eh.sendWebhook(webhookConfig, eventType, eventData, client.SessionID)
}

func (eh *DefaultEventHandler) shouldSendWebhook(webhookConfig *webhook.Webhook, eventType EventType) bool {
	log.Debug().
		Str("pkg", "waclient").
		Bool("webhook_sender_nil", eh.webhookSender == nil).
		Bool("webhook_enabled", webhookConfig.Enabled).
		Str("event_type", string(eventType)).
		Interface("webhook_events", webhookConfig.Events).
		Msg("DEBUG: shouldSendWebhook evaluation")

	if eh.webhookSender == nil {
		log.Debug().
			Str("pkg", "waclient").
			Msg("DEBUG: webhookSender is nil")
		return false
	}

	if !webhookConfig.Enabled {
		log.Debug().
			Str("pkg", "waclient").
			Msg("DEBUG: webhook is disabled")
		return false
	}

	if len(webhookConfig.Events) == 0 {
		log.Debug().
			Str("pkg", "waclient").
			Msg("DEBUG: no events configured, allowing all")
		return true
	}

	eventTypeStr := string(eventType)
	for _, subscribedEvent := range webhookConfig.Events {
		if subscribedEvent == eventTypeStr {
			log.Debug().
				Str("pkg", "waclient").
				Str("event_type", eventTypeStr).
				Msg("DEBUG: event type matches subscription")
			return true
		}
	}

	log.Debug().
		Str("pkg", "waclient").
		Str("event_type", eventTypeStr).
		Msg("DEBUG: event type does not match any subscription")
	return false
}

// OrderedWebhookPayload garante a ordem específica dos campos no JSON
type OrderedWebhookPayload struct {
	Event     string      `json:"event"`
	SessionID string      `json:"sessionId"`
	Timestamp int64       `json:"timestamp"`
	Data      interface{} `json:"data"`
}

func (eh *DefaultEventHandler) sendWebhook(webhookConfig *webhook.Webhook, eventType EventType, eventData interface{}, sessionID string) error {

	// Criar payload com ordem específica: event, sessionId, timestamp, data
	orderedPayload := OrderedWebhookPayload{
		Event:     string(eventType),
		SessionID: sessionID,
		Timestamp: time.Now().Unix(),
		Data:      eventData,
	}

	// Marcar que este payload deve ser serializado diretamente
	payloadMap := map[string]interface{}{
		"_ordered_payload": orderedPayload,
	}

	webhookEvent := &output.WebhookEvent{
		ID:        uuid.New().String(),
		Type:      string(eventType),
		SessionID: sessionID,
		Timestamp: time.Now(),
		Data:      payloadMap,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return eh.webhookSender.SendWebhook(ctx, webhookConfig.URL, webhookConfig.Secret, webhookEvent)
}

func getMessageType(msg interface{}) string {
	msgMap, ok := convertMessageToMap(msg)
	if !ok {
		return unknownMessageType
	}
	return detectMessageType(msgMap)
}

func convertMessageToMap(msg interface{}) (map[string]interface{}, bool) {
	if msg == nil {
		return nil, false
	}

	msgJSON, err := json.Marshal(msg)
	if err != nil {
		return nil, false
	}

	var msgMap map[string]interface{}
	if err := json.Unmarshal(msgJSON, &msgMap); err != nil {
		return nil, false
	}
	return msgMap, true
}

func detectMessageType(msgMap map[string]interface{}) string {
	if hasMessageField(msgMap, "conversation", "extendedTextMessage") {
		return "text"
	}
	if hasMessageField(msgMap, "imageMessage") {
		return "image"
	}
	if hasMessageField(msgMap, "audioMessage") {
		return "audio"
	}
	if hasMessageField(msgMap, "videoMessage") {
		return "video"
	}
	if hasMessageField(msgMap, "documentMessage") {
		return "document"
	}
	if hasMessageField(msgMap, "stickerMessage") {
		return "sticker"
	}
	if hasMessageField(msgMap, "locationMessage") {
		return "location"
	}
	if hasMessageField(msgMap, "liveLocationMessage") {
		return "liveLocation"
	}
	if hasMessageField(msgMap, "contactMessage") {
		return "contact"
	}
	if hasMessageField(msgMap, "contactsArrayMessage") {
		return "contacts"
	}
	if hasMessageField(msgMap, "buttonsMessage") {
		return "buttons"
	}
	if hasMessageField(msgMap, "listMessage") {
		return "list"
	}
	if hasMessageField(msgMap, "templateMessage") {
		return "template"
	}
	return unknownMessageType
}

func hasMessageField(msgMap map[string]interface{}, fields ...string) bool {
	for _, field := range fields {
		if _, ok := msgMap[field]; ok {
			return true
		}
	}
	return false
}
