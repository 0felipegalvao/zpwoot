package wmeow

import (
	"context"
	"time"

	"zpwoot/platform/logger"

	"go.mau.fi/whatsmeow/types/events"
)

// EventHandler handles WhatsApp events
type EventHandler struct {
	manager    *Manager
	sessionMgr *SessionManager
	qrGen      *QRCodeGenerator
	logger     *logger.Logger
}

// NewEventHandler creates a new event handler
func NewEventHandler(manager *Manager, sessionMgr *SessionManager, qrGen *QRCodeGenerator, logger *logger.Logger) *EventHandler {
	return &EventHandler{
		manager:    manager,
		sessionMgr: sessionMgr,
		qrGen:      qrGen,
		logger:     logger,
	}
}

// SetupEventHandlers is now defined in manager.go to avoid circular imports

// HandleEvent handles all WhatsApp events
func (h *EventHandler) HandleEvent(evt interface{}, sessionID string) {
	switch v := evt.(type) {
	case *events.Connected:
		h.handleConnected(v, sessionID)
	case *events.Disconnected:
		h.handleDisconnected(v, sessionID)
	case *events.LoggedOut:
		h.handleLoggedOut(v, sessionID)
	case *events.QR:
		h.handleQR(v, sessionID)
	case *events.PairSuccess:
		h.handlePairSuccess(v, sessionID)
	case *events.PairError:
		h.handlePairError(v, sessionID)
	case *events.Message:
		h.handleMessage(v, sessionID)
	case *events.Receipt:
		h.handleReceipt(v, sessionID)
	case *events.Presence:
		h.handlePresence(v, sessionID)
	case *events.ChatPresence:
		h.handleChatPresence(v, sessionID)
	case *events.HistorySync:
		h.handleHistorySync(v, sessionID)
	default:
		h.logger.InfoWithFields("Unhandled event", map[string]interface{}{
			"session_id": sessionID,
			"event_type": getEventType(evt),
		})
	}
}

// handleConnected handles connection events
func (h *EventHandler) handleConnected(evt *events.Connected, sessionID string) {
	h.logger.InfoWithFields("WhatsApp connected", map[string]interface{}{
		"session_id":   sessionID,
		"event_type":   "Connected",
		"connected_at": time.Now().Unix(),
	})

	// Use evt to avoid unused parameter warning
	_ = evt

	h.sessionMgr.UpdateConnectionStatus(sessionID, true)
}

// handleDisconnected handles disconnection events
func (h *EventHandler) handleDisconnected(evt *events.Disconnected, sessionID string) {
	h.logger.InfoWithFields("WhatsApp disconnected", map[string]interface{}{
		"session_id":      sessionID,
		"event_type":      "Disconnected",
		"disconnected_at": time.Now().Unix(),
	})

	// Use evt to avoid unused parameter warning
	_ = evt

	h.sessionMgr.UpdateConnectionStatus(sessionID, false)
}

// handleLoggedOut handles logout events
func (h *EventHandler) handleLoggedOut(evt *events.LoggedOut, sessionID string) {
	h.logger.InfoWithFields("WhatsApp logged out", map[string]interface{}{
		"session_id": sessionID,
		"reason":     evt.Reason,
	})

	h.sessionMgr.UpdateConnectionStatus(sessionID, false)
}

// handleQR handles QR code events
func (h *EventHandler) handleQR(evt *events.QR, sessionID string) {
	h.logger.InfoWithFields("QR code received", map[string]interface{}{
		"session_id":  sessionID,
		"codes_count": len(evt.Codes),
	})

	// Generate QR code image
	qrImage := h.qrGen.GenerateQRCodeImage(evt.Codes[0])

	// Update session with QR code
	h.updateSessionQRCode(sessionID, qrImage)

	// Display QR code in terminal
	h.qrGen.DisplayQRCodeInTerminal(evt.Codes[0], sessionID)
}

// handlePairSuccess handles successful pairing
func (h *EventHandler) handlePairSuccess(evt *events.PairSuccess, sessionID string) {
	h.logger.InfoWithFields("Pairing successful", map[string]interface{}{
		"session_id": sessionID,
		"device_jid": evt.ID.String(),
	})

	h.sessionMgr.UpdateConnectionStatus(sessionID, true)

	// Update session with device JID
	h.updateSessionDeviceJID(sessionID, evt.ID.String())
}

// handlePairError handles pairing errors
func (h *EventHandler) handlePairError(evt *events.PairError, sessionID string) {
	h.logger.ErrorWithFields("Pairing failed", map[string]interface{}{
		"session_id": sessionID,
		"error":      evt.Error.Error(),
	})

	h.sessionMgr.UpdateConnectionStatus(sessionID, false)
}

// handleMessage handles incoming messages
func (h *EventHandler) handleMessage(evt *events.Message, sessionID string) {
	h.logger.InfoWithFields("Message received", map[string]interface{}{
		"session_id": sessionID,
		"from":       evt.Info.Sender.String(),
		"message_id": evt.Info.ID,
		"timestamp":  evt.Info.Timestamp,
	})

	// Update last seen
	h.updateSessionLastSeen(sessionID)

	// Here you would typically:
	// 1. Process the message
	// 2. Send to webhooks
	// 3. Forward to Chatwoot if configured
	// 4. Store in database if needed
}

// handleReceipt handles message receipts
func (h *EventHandler) handleReceipt(evt *events.Receipt, sessionID string) {
	h.logger.InfoWithFields("Receipt received", map[string]interface{}{
		"session_id": sessionID,
		"type":       evt.Type,
		"sender":     evt.Sender.String(),
		"timestamp":  evt.Timestamp,
	})
}

// handlePresence handles presence updates
func (h *EventHandler) handlePresence(evt *events.Presence, sessionID string) {
	h.logger.InfoWithFields("Presence update", map[string]interface{}{
		"session_id":  sessionID,
		"from":        evt.From.String(),
		"unavailable": evt.Unavailable,
		"last_seen":   evt.LastSeen,
	})
}

// handleChatPresence handles chat presence updates
func (h *EventHandler) handleChatPresence(evt *events.ChatPresence, sessionID string) {
	h.logger.InfoWithFields("Chat presence update", map[string]interface{}{
		"session_id": sessionID,
		"chat":       evt.Chat.String(),
		"state":      evt.State,
	})
}

// handleHistorySync handles history sync events
func (h *EventHandler) handleHistorySync(evt *events.HistorySync, sessionID string) {
	h.logger.InfoWithFields("History sync", map[string]interface{}{
		"session_id": sessionID,
		"data_size":  len(evt.Data.String()), // Just log the data size for now
	})
}

// updateSessionQRCode updates the QR code for a session
func (h *EventHandler) updateSessionQRCode(sessionID, qrCode string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sess, err := h.sessionMgr.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		h.logger.ErrorWithFields("Failed to get session for QR update", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return
	}

	// Update QR code
	sess.QRCode = qrCode
	sess.UpdatedAt = time.Now()

	if err := h.sessionMgr.sessionRepo.Update(ctx, sess); err != nil {
		h.logger.ErrorWithFields("Failed to update session QR code", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
	}
}

// updateSessionDeviceJID updates the device JID for a session
func (h *EventHandler) updateSessionDeviceJID(sessionID, deviceJID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sess, err := h.sessionMgr.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		h.logger.ErrorWithFields("Failed to get session for device JID update", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return
	}

	sess.DeviceJid = deviceJID
	sess.UpdatedAt = time.Now()

	if err := h.sessionMgr.sessionRepo.Update(ctx, sess); err != nil {
		h.logger.ErrorWithFields("Failed to update session device JID", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
	}
}

// updateSessionLastSeen updates the last seen timestamp for a session
func (h *EventHandler) updateSessionLastSeen(sessionID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sess, err := h.sessionMgr.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		h.logger.ErrorWithFields("Failed to get session for last seen update", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return
	}

	now := time.Now()
	sess.LastSeen = &now
	sess.UpdatedAt = now

	if err := h.sessionMgr.sessionRepo.Update(ctx, sess); err != nil {
		h.logger.ErrorWithFields("Failed to update session last seen", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
	}
}

// getEventType returns the type name of an event
func getEventType(evt interface{}) string {
	switch evt.(type) {
	case *events.Connected:
		return "Connected"
	case *events.Disconnected:
		return "Disconnected"
	case *events.LoggedOut:
		return "LoggedOut"
	case *events.QR:
		return "QR"
	case *events.PairSuccess:
		return "PairSuccess"
	case *events.PairError:
		return "PairError"
	case *events.Message:
		return "Message"
	case *events.Receipt:
		return "Receipt"
	case *events.Presence:
		return "Presence"
	case *events.ChatPresence:
		return "ChatPresence"
	case *events.HistorySync:
		return "HistorySync"
	default:
		return "Unknown"
	}
}
