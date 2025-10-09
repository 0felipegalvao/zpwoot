package session

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID              string     `json:"id" db:"id"`
	Name            string     `json:"name" db:"name"`
	APIKey          string     `json:"apiKey" db:"apiKey"`
	DeviceJID       string     `json:"device_jid,omitempty" db:"deviceJid"`
	IsConnected     bool       `json:"is_connected" db:"isConnected"`
	ConnectionError string     `json:"connection_error,omitempty" db:"connectionError"`
	QRCode          string     `json:"qr_code,omitempty" db:"qrCode"`
	QRCodeExpiresAt *time.Time `json:"qr_code_expires_at,omitempty" db:"qrCodeExpiresAt"`
	CreatedAt       time.Time  `json:"created_at" db:"createdAt"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updatedAt"`
}

type Status string

const (
	StatusDisconnected Status = "disconnected"
	StatusConnecting   Status = "connecting"
	StatusConnected    Status = "connected"
	StatusError        Status = "error"
)

func (s Status) IsValid() bool {
	switch s {
	case StatusDisconnected, StatusConnecting, StatusConnected, StatusError:
		return true
	default:
		return false
	}
}

func GenerateAPIKey() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return strings.ToUpper(strings.ReplaceAll(uuid.New().String(), "-", ""))[:32]
	}
	return strings.ToUpper(hex.EncodeToString(bytes))[:32]
}

func NewSession(name string) *Session {
	now := time.Now()
	return &Session{
		ID:          uuid.New().String(),
		Name:        name,
		APIKey:      GenerateAPIKey(),
		IsConnected: false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func (s *Session) GetStatus() Status {
	if s.IsConnected {
		return StatusConnected
	}
	if s.QRCode != "" && s.isQRCodeValid() {
		return StatusConnecting
	}
	if s.ConnectionError != "" {
		return StatusError
	}
	return StatusDisconnected
}

func (s *Session) SetQRCode(qrCode string, expiresAt time.Time) {
	s.QRCode = qrCode
	s.QRCodeExpiresAt = &expiresAt
	s.touch()
}

func (s *Session) ClearQRCode() {
	s.QRCode = ""
	s.QRCodeExpiresAt = nil
	s.touch()
}

func (s *Session) SetConnected(deviceJID string) {
	s.IsConnected = true
	s.DeviceJID = deviceJID
	s.ConnectionError = ""
	s.ClearQRCode()
	s.touch()
}

func (s *Session) SetDisconnected() {
	s.IsConnected = false
	s.ClearQRCode()
	s.touch()
}

func (s *Session) SetError(errMsg string) {
	s.IsConnected = false
	s.ConnectionError = errMsg
	s.ClearQRCode()
	s.touch()
}

func (s *Session) isQRCodeValid() bool {
	return s.QRCodeExpiresAt == nil || time.Now().Before(*s.QRCodeExpiresAt)
}

func (s *Session) touch() {
	s.UpdatedAt = time.Now()
}
