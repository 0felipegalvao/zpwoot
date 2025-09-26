package wameow

import (
	"context"
	"fmt"
	"sync"
	"time"

	"zpwoot/internal/ports"
	"zpwoot/platform/logger"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waTypes "go.mau.fi/whatsmeow/types"
)

// WameowClient wraps whatsmeow.Client with additional functionality
type WameowClient struct {
	sessionID   string
	client      *whatsmeow.Client
	logger      *logger.Logger
	sessionMgr  *SessionManager
	qrGenerator *QRCodeGenerator

	mu           sync.RWMutex
	status       string
	lastActivity time.Time

	qrCode       string
	qrCodeBase64 string
	qrLoopActive bool

	ctx           context.Context
	cancel        context.CancelFunc
	qrStopChannel chan bool
}

// NewWameowClient creates a new WameowClient
func NewWameowClient(
	sessionID string,
	container *sqlstore.Container,
	sessionRepo ports.SessionRepository,
	logger *logger.Logger,
) (*WameowClient, error) {
	// Get session from repository to check for existing deviceJid
	ctx := context.Background()
	sess, err := sessionRepo.GetByID(ctx, sessionID)
	var deviceJid string
	if err == nil && sess != nil {
		deviceJid = sess.DeviceJid
		logger.InfoWithFields("Found existing session", map[string]interface{}{
			"session_id": sessionID,
			"device_jid": deviceJid,
		})
	} else {
		logger.InfoWithFields("Creating new session", map[string]interface{}{
			"session_id": sessionID,
		})
	}

	// Get device store for session with the correct deviceJid
	deviceStore := GetDeviceStoreForSession(sessionID, deviceJid, container)
	if deviceStore == nil {
		return nil, fmt.Errorf("failed to create device store for session %s", sessionID)
	}

	// Create whatsmeow logger
	waLogger := NewWameowLogger(logger)

	// Create whatsmeow client
	client := whatsmeow.NewClient(deviceStore, waLogger)
	if client == nil {
		return nil, fmt.Errorf("failed to create WhatsApp client for session %s", sessionID)
	}

	ctx, cancel := context.WithCancel(context.Background())

	wameowClient := &WameowClient{
		sessionID:     sessionID,
		client:        client,
		logger:        logger,
		sessionMgr:    NewSessionManager(sessionRepo, logger),
		qrGenerator:   NewQRCodeGenerator(logger),
		status:        "disconnected",
		lastActivity:  time.Now(),
		ctx:           ctx,
		cancel:        cancel,
		qrStopChannel: make(chan bool, 1),
	}

	return wameowClient, nil
}

// Connect starts the connection process
func (c *WameowClient) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.logger.InfoWithFields("Starting connection process (will restart if already running)", map[string]interface{}{
		"session_id": c.sessionID,
	})

	// Always stop any existing QR loop first
	c.stopQRLoop()

	// If client is connected, disconnect first to restart the process
	if c.client.IsConnected() {
		c.logger.InfoWithFields("Client already connected, disconnecting to restart", map[string]interface{}{
			"session_id": c.sessionID,
		})
		c.client.Disconnect()
	}

	// Cancel any existing context and create a new one
	if c.cancel != nil {
		c.cancel()
	}
	c.ctx, c.cancel = context.WithCancel(context.Background())

	c.setStatus("connecting")

	// Start connection process in background
	go c.startClientLoop()

	return nil
}

// Disconnect stops the connection
func (c *WameowClient) Disconnect() error {
	c.logger.InfoWithFields("Disconnecting client", map[string]interface{}{
		"session_id": c.sessionID,
	})

	c.mu.Lock()
	defer c.mu.Unlock()

	c.stopQRLoop()

	if c.client.IsConnected() {
		c.client.Disconnect()
	}

	if c.cancel != nil {
		c.cancel()
	}

	c.setStatus("disconnected")
	return nil
}

// IsConnected returns connection status
func (c *WameowClient) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.client.IsConnected()
}

// IsLoggedIn returns login status
func (c *WameowClient) IsLoggedIn() bool {
	return c.client.IsLoggedIn()
}

// GetQRCode returns the current QR code
func (c *WameowClient) GetQRCode() (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.qrCode == "" {
		return "", fmt.Errorf("no QR code available")
	}

	return c.qrCode, nil
}

// GetClient returns the underlying whatsmeow client
func (c *WameowClient) GetClient() *whatsmeow.Client {
	return c.client
}

// GetJID returns the device JID
func (c *WameowClient) GetJID() waTypes.JID {
	if c.client.Store.ID == nil {
		return waTypes.EmptyJID
	}
	return *c.client.Store.ID
}

// setStatus sets the internal status
func (c *WameowClient) setStatus(status string) {
	c.status = status
	c.lastActivity = time.Now()
	c.logger.InfoWithFields("Session status updated", map[string]interface{}{
		"session_id": c.sessionID,
		"status":     status,
	})

	// Update database when status changes to connected or disconnected
	if status == "connected" {
		c.sessionMgr.UpdateConnectionStatus(c.sessionID, true)
	} else if status == "disconnected" {
		c.sessionMgr.UpdateConnectionStatus(c.sessionID, false)
	}
}

// startClientLoop handles the connection logic
func (c *WameowClient) startClientLoop() {
	defer func() {
		if r := recover(); r != nil {
			c.logger.ErrorWithFields("Client loop panic", map[string]interface{}{
				"session_id": c.sessionID,
				"error":      r,
			})
		}
	}()

	if !IsDeviceRegistered(c.client) {
		c.logger.InfoWithFields("Device not registered, starting QR code process", map[string]interface{}{
			"session_id": c.sessionID,
		})
		c.handleNewDeviceRegistration()
	} else {
		c.logger.InfoWithFields("Device already registered, connecting directly", map[string]interface{}{
			"session_id": c.sessionID,
		})
		c.handleExistingDeviceConnection()
	}
}

// handleNewDeviceRegistration handles QR code generation for new devices
func (c *WameowClient) handleNewDeviceRegistration() {
	qrChan, err := c.client.GetQRChannel(context.Background())
	if err != nil {
		c.logger.ErrorWithFields("Failed to get QR channel", map[string]interface{}{
			"session_id": c.sessionID,
			"error":      err.Error(),
		})
		c.setStatus("disconnected")
		return
	}

	err = c.client.Connect()
	if err != nil {
		c.logger.ErrorWithFields("Failed to connect client", map[string]interface{}{
			"session_id": c.sessionID,
			"error":      err.Error(),
		})
		c.setStatus("disconnected")
		return
	}

	c.handleQRLoop(qrChan)
}

// handleExistingDeviceConnection handles connection for registered devices
func (c *WameowClient) handleExistingDeviceConnection() {
	err := c.client.Connect()
	if err != nil {
		c.logger.ErrorWithFields("Failed to connect existing device", map[string]interface{}{
			"session_id": c.sessionID,
			"error":      err.Error(),
		})
		c.setStatus("disconnected")
		return
	}

	time.Sleep(2 * time.Second)

	if c.client.IsConnected() {
		c.logger.InfoWithFields("Successfully connected session", map[string]interface{}{
			"session_id": c.sessionID,
		})
		c.setStatus("connected")
	} else {
		c.logger.WarnWithFields("Connection attempt completed but client not connected", map[string]interface{}{
			"session_id": c.sessionID,
		})
		c.setStatus("disconnected")
	}
}

// handleQRLoop handles QR code events
func (c *WameowClient) handleQRLoop(qrChan <-chan whatsmeow.QRChannelItem) {
	if qrChan == nil {
		c.logger.ErrorWithFields("QR channel is nil", map[string]interface{}{
			"session_id": c.sessionID,
		})
		return
	}

	c.mu.Lock()
	c.qrLoopActive = true
	c.mu.Unlock()

	defer func() {
		if r := recover(); r != nil {
			c.logger.ErrorWithFields("QR loop panic", map[string]interface{}{
				"session_id": c.sessionID,
				"error":      r,
			})
		}
		c.mu.Lock()
		c.qrLoopActive = false
		c.mu.Unlock()
	}()

	for {
		select {
		case <-c.ctx.Done():
			c.logger.InfoWithFields("QR loop cancelled", map[string]interface{}{
				"session_id": c.sessionID,
			})
			return

		case <-c.qrStopChannel:
			c.logger.InfoWithFields("QR loop stopped", map[string]interface{}{
				"session_id": c.sessionID,
			})
			return

		case evt, ok := <-qrChan:
			if !ok {
				c.logger.InfoWithFields("QR channel closed", map[string]interface{}{
					"session_id": c.sessionID,
				})
				c.setStatus("disconnected")
				return
			}

			switch evt.Event {
			case "code":
				c.mu.Lock()
				c.qrCode = evt.Code
				if c.qrGenerator != nil {
					c.qrCodeBase64 = c.qrGenerator.GenerateQRCodeImage(evt.Code)
				}
				c.mu.Unlock()

				// Display compact QR code in terminal (only once)
				if c.qrGenerator != nil {
					c.qrGenerator.DisplayQRCodeInTerminal(evt.Code, c.sessionID)
				}

				c.logger.InfoWithFields("QR code generated", map[string]interface{}{
					"session_id": c.sessionID,
				})
				c.setStatus("connecting")

			case "success":
				c.logger.InfoWithFields("QR code scanned successfully", map[string]interface{}{
					"session_id": c.sessionID,
				})
				c.setStatus("connected")
				return

			case "timeout":
				c.logger.WarnWithFields("QR code timeout", map[string]interface{}{
					"session_id": c.sessionID,
				})
				c.mu.Lock()
				c.qrCode = ""
				c.qrCodeBase64 = ""
				c.mu.Unlock()
				c.setStatus("disconnected")
				return

			default:
				c.logger.InfoWithFields("QR event", map[string]interface{}{
					"session_id": c.sessionID,
					"event":      evt.Event,
				})
			}
		}
	}
}

// stopQRLoop stops the QR code loop
func (c *WameowClient) stopQRLoop() {
	if c.qrLoopActive {
		c.logger.InfoWithFields("Stopping existing QR loop", map[string]interface{}{
			"session_id": c.sessionID,
		})
		select {
		case c.qrStopChannel <- true:
			c.logger.InfoWithFields("QR loop stop signal sent", map[string]interface{}{
				"session_id": c.sessionID,
			})
		default:
			c.logger.InfoWithFields("QR loop stop channel full, loop may already be stopping", map[string]interface{}{
				"session_id": c.sessionID,
			})
		}
		// Wait a bit for the loop to stop
		time.Sleep(100 * time.Millisecond)
	}
}

// Logout logs out the session
func (c *WameowClient) Logout() error {
	c.logger.InfoWithFields("Logging out session", map[string]interface{}{
		"session_id": c.sessionID,
	})

	err := c.client.Logout(context.Background())
	if err != nil {
		c.logger.ErrorWithFields("Failed to logout session", map[string]interface{}{
			"session_id": c.sessionID,
			"error":      err.Error(),
		})
		return fmt.Errorf("failed to logout: %w", err)
	}

	if c.client.IsConnected() {
		c.client.Disconnect()
	}

	c.setStatus("disconnected")
	c.logger.InfoWithFields("Successfully logged out session", map[string]interface{}{
		"session_id": c.sessionID,
	})
	return nil
}

// SendMessage sends a message (placeholder for now)
func (c *WameowClient) SendMessage(ctx context.Context, to string, message interface{}) (*whatsmeow.SendResponse, error) {
	// This is a placeholder - in a real implementation, you would handle different message types
	return nil, fmt.Errorf("SendMessage not implemented yet")
}

// Upload uploads media (placeholder for now)
func (c *WameowClient) Upload(ctx context.Context, data []byte, appInfo whatsmeow.MediaType) (*whatsmeow.UploadResponse, error) {
	// This is a placeholder - in a real implementation, you would handle media upload
	return nil, fmt.Errorf("Upload not implemented yet")
}

// AddEventHandler adds an event handler
func (c *WameowClient) AddEventHandler(handler whatsmeow.EventHandler) uint32 {
	return c.client.AddEventHandler(handler)
}

// IsDeviceRegistered checks if the device is registered (has a store ID)
func IsDeviceRegistered(client *whatsmeow.Client) bool {
	if client == nil || client.Store == nil {
		return false
	}
	return client.Store.ID != nil
}
