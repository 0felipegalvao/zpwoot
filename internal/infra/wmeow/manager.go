package wmeow

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"zpwoot/internal/domain/session"
	"zpwoot/internal/ports"
	"zpwoot/platform/logger"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

// SessionStats tracks statistics for a session
type SessionStats struct {
	MessagesSent     int64
	MessagesReceived int64
	LastActivity     int64
	StartTime        int64
}

// EventHandlerInfo stores information about registered event handlers
type EventHandlerInfo struct {
	ID      string
	Handler ports.EventHandler
}

// Manager implements the WhatsAppManager interface
type Manager struct {
	clients       map[string]*whatsmeow.Client
	clientsMutex  sync.RWMutex
	container     *sqlstore.Container
	connectionMgr *ConnectionManager
	qrGenerator   *QRCodeGenerator
	sessionMgr    *SessionManager
	logger        *logger.Logger

	// Statistics tracking
	sessionStats map[string]*SessionStats
	statsMutex   sync.RWMutex

	// Event handlers
	eventHandlers map[string]map[string]*EventHandlerInfo // sessionID -> handlerID -> handler
	handlersMutex sync.RWMutex
}

// NewManager creates a new WhatsApp manager
func NewManager(
	container *sqlstore.Container,
	sessionRepo ports.SessionRepository,
	logger *logger.Logger,
) *Manager {
	return &Manager{
		clients:       make(map[string]*whatsmeow.Client),
		container:     container,
		connectionMgr: NewConnectionManager(logger),
		qrGenerator:   NewQRCodeGenerator(logger),
		sessionMgr:    NewSessionManager(sessionRepo, logger),
		logger:        logger,
	}
}

// CreateSession creates a new WhatsApp session
func (m *Manager) CreateSession(sessionID string, config *session.ProxyConfig) error {
	m.logger.InfoWithFields("Creating WhatsApp session", map[string]interface{}{
		"session_id": sessionID,
	})

	m.clientsMutex.Lock()
	defer m.clientsMutex.Unlock()

	// Check if session already exists
	if _, exists := m.clients[sessionID]; exists {
		return fmt.Errorf("session %s already exists", sessionID)
	}

	// Get device store for session
	deviceStore := GetDeviceStoreForSession(sessionID, "", m.container)
	if deviceStore == nil {
		return fmt.Errorf("failed to create device store for session %s", sessionID)
	}

	// Create WhatsApp logger wrapper
	waLogger := NewWhatsAppLogger(m.logger)

	// Create WhatsApp client
	client := whatsmeow.NewClient(deviceStore, waLogger)
	if client == nil {
		return fmt.Errorf("failed to create WhatsApp client for session %s", sessionID)
	}

	// Set up event handlers
	m.setupEventHandlers(client, sessionID)

	// Apply proxy configuration if provided
	if config != nil {
		if err := m.applyProxyConfig(client, config); err != nil {
			m.logger.WarnWithFields("Failed to apply proxy config", map[string]interface{}{
				"session_id": sessionID,
				"error":      err.Error(),
			})
		}
	}

	// Store client
	m.clients[sessionID] = client

	m.logger.InfoWithFields("WhatsApp session created successfully", map[string]interface{}{
		"session_id": sessionID,
	})

	return nil
}

// ConnectSession connects a WhatsApp session
func (m *Manager) ConnectSession(sessionID string) error {
	m.logger.InfoWithFields("Connecting WhatsApp session", map[string]interface{}{
		"session_id": sessionID,
	})

	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("session %s not found", sessionID)
	}

	// Update session status to connecting
	// Connection status will be updated by event handlers

	// Connect with retry
	config := &RetryConfig{
		MaxRetries:    3,
		RetryInterval: 10 * time.Second,
	}

	err := m.connectionMgr.ConnectWithRetry(client, sessionID, config)
	if err != nil {
		m.sessionMgr.UpdateConnectionStatus(sessionID, false)
		return fmt.Errorf("failed to connect session %s: %w", sessionID, err)
	}

	return nil
}

// DisconnectSession disconnects a WhatsApp session
func (m *Manager) DisconnectSession(sessionID string) error {
	m.logger.InfoWithFields("Disconnecting WhatsApp session", map[string]interface{}{
		"session_id": sessionID,
	})

	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("session %s not found", sessionID)
	}

	m.connectionMgr.SafeDisconnect(client, sessionID)
	m.sessionMgr.UpdateConnectionStatus(sessionID, false)

	return nil
}

// LogoutSession logs out a WhatsApp session
func (m *Manager) LogoutSession(sessionID string) error {
	m.logger.InfoWithFields("Logging out WhatsApp session", map[string]interface{}{
		"session_id": sessionID,
	})

	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("session %s not found", sessionID)
	}

	// Logout from WhatsApp
	ctx := context.Background()
	err := client.Logout(ctx)
	if err != nil {
		m.logger.WarnWithFields("Error during logout", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
	}

	// Update session status
	m.sessionMgr.UpdateConnectionStatus(sessionID, false)

	// Remove client from memory
	m.clientsMutex.Lock()
	delete(m.clients, sessionID)
	m.clientsMutex.Unlock()

	return nil
}

// GetQRCode gets QR code for session pairing
func (m *Manager) GetQRCode(sessionID string) (*session.QRCodeResponse, error) {
	m.logger.InfoWithFields("Getting QR code for session", map[string]interface{}{
		"session_id": sessionID,
	})

	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}

	if client.IsLoggedIn() {
		return nil, fmt.Errorf("session %s is already logged in", sessionID)
	}

	// This would typically be handled by the QR event handler
	// For now, return a placeholder response
	return &session.QRCodeResponse{
		QRCode:    "placeholder_qr_code",
		ExpiresAt: time.Now().Add(2 * time.Minute),
		Timeout:   120,
	}, nil
}

// PairPhone pairs a phone number with the session
func (m *Manager) PairPhone(sessionID, phoneNumber string) error {
	m.logger.InfoWithFields("Pairing phone number", map[string]interface{}{
		"session_id":   sessionID,
		"phone_number": phoneNumber,
	})

	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("session %s not found", sessionID)
	}

	// This would implement phone pairing logic
	// For now, return not implemented
	return fmt.Errorf("phone pairing not implemented yet")
}

// IsConnected checks if a session is connected
func (m *Manager) IsConnected(sessionID string) bool {
	client := m.getClient(sessionID)
	if client == nil {
		return false
	}
	return client.IsConnected()
}

// GetDeviceInfo gets device information for a session
func (m *Manager) GetDeviceInfo(sessionID string) (*session.DeviceInfo, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}

	if !client.IsLoggedIn() {
		return nil, fmt.Errorf("session %s is not logged in", sessionID)
	}

	// This would get actual device info from WhatsApp
	// For now, return placeholder data
	return &session.DeviceInfo{
		Platform:    "web",
		DeviceModel: "Chrome",
		OSVersion:   "Unknown",
		AppVersion:  "2.2412.54",
	}, nil
}

// SetProxy sets proxy configuration for a session
func (m *Manager) SetProxy(sessionID string, config *session.ProxyConfig) error {
	m.logger.InfoWithFields("Setting proxy for session", map[string]interface{}{
		"session_id": sessionID,
		"proxy_type": config.Type,
		"proxy_host": config.Host,
	})

	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("session %s not found", sessionID)
	}

	return m.applyProxyConfig(client, config)
}

// GetProxy gets proxy configuration for a session
func (m *Manager) GetProxy(sessionID string) (*session.ProxyConfig, error) {
	// This would get the current proxy configuration
	// For now, return nil (no proxy)
	return nil, nil
}

// GetSessionStats retrieves session statistics
func (m *Manager) GetSessionStats(sessionID string) (*ports.SessionStats, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}

	// Return basic session statistics
	return &ports.SessionStats{
		MessagesSent:     0, // TODO: implement message counting
		MessagesReceived: 0, // TODO: implement message counting
		LastActivity:     time.Now().Unix(),
		Uptime:           0, // TODO: implement uptime calculation
	}, nil
}

// SendMessage sends a message through WhatsApp
func (m *Manager) SendMessage(sessionID, to, message string) error {
	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("session %s not found", sessionID)
	}

	if !client.IsLoggedIn() {
		return fmt.Errorf("session %s is not logged in", sessionID)
	}

	// TODO: implement message sending
	return fmt.Errorf("message sending not implemented yet")
}

// SendMediaMessage sends a media message
func (m *Manager) SendMediaMessage(sessionID, to string, media []byte, mediaType, caption string) error {
	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("session %s not found", sessionID)
	}

	if !client.IsLoggedIn() {
		return fmt.Errorf("session %s is not logged in", sessionID)
	}

	// TODO: implement media message sending
	return fmt.Errorf("media message sending not implemented yet")
}

// RegisterEventHandler registers an event handler for WhatsApp events
func (m *Manager) RegisterEventHandler(sessionID string, handler ports.EventHandler) error {
	// TODO: implement event handler registration
	return fmt.Errorf("event handler registration not implemented yet")
}

// UnregisterEventHandler removes an event handler
func (m *Manager) UnregisterEventHandler(sessionID string, handlerID string) error {
	// TODO: implement event handler unregistration
	return fmt.Errorf("event handler unregistration not implemented yet")
}

// getClient safely gets a client by session ID
func (m *Manager) getClient(sessionID string) *whatsmeow.Client {
	m.clientsMutex.RLock()
	defer m.clientsMutex.RUnlock()
	return m.clients[sessionID]
}

// applyProxyConfig applies proxy configuration to a client
func (m *Manager) applyProxyConfig(client *whatsmeow.Client, config *session.ProxyConfig) error {
	// This would implement proxy configuration
	// For now, just log the configuration and validate client
	m.logger.InfoWithFields("Proxy configuration", map[string]interface{}{
		"type":       config.Type,
		"host":       config.Host,
		"port":       config.Port,
		"client_nil": client == nil,
	})

	if client == nil {
		return fmt.Errorf("cannot apply proxy config to nil client")
	}

	// TODO: Implement actual proxy configuration
	// This would typically involve setting up HTTP/SOCKS proxy for the client

	return nil
}

// setupEventHandlers sets up event handlers for a WhatsApp client
func (m *Manager) setupEventHandlers(client *whatsmeow.Client, sessionID string) {
	m.logger.InfoWithFields("Setting up event handlers", map[string]interface{}{
		"session_id": sessionID,
	})

	// Set up the actual event handlers
	m.SetupEventHandlers(client, sessionID)
}

// SetupEventHandlers sets up all event handlers for a WhatsApp client
func (m *Manager) SetupEventHandlers(client *whatsmeow.Client, sessionID string) {
	eventHandler := NewEventHandler(m, m.sessionMgr, m.qrGenerator, m.logger)

	client.AddEventHandler(func(evt interface{}) {
		eventHandler.HandleEvent(evt, sessionID)
	})
}
