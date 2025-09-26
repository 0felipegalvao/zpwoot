package session

import (
	"context"
	"time"

	"zpwoot/pkg/errors"
	"zpwoot/pkg/uuid"
)

type Service struct {
	repo      Repository
	whatsapp  WhatsAppManager
	generator *uuid.Generator
}

type Repository interface {
	Create(ctx context.Context, session *Session) error
	GetByID(ctx context.Context, id string) (*Session, error)
	GetByDeviceJid(ctx context.Context, deviceJid string) (*Session, error)
	List(ctx context.Context, req *ListSessionsRequest) ([]*Session, int, error)
	Update(ctx context.Context, session *Session) error
	Delete(ctx context.Context, id string) error
}

type WhatsAppManager interface {
	CreateSession(sessionID string, config *ProxyConfig) error
	ConnectSession(sessionID string) error
	DisconnectSession(sessionID string) error
	LogoutSession(sessionID string) error
	GetQRCode(sessionID string) (*QRCodeResponse, error)
	PairPhone(sessionID, phoneNumber string) error
	IsConnected(sessionID string) bool
	GetDeviceInfo(sessionID string) (*DeviceInfo, error)
	SetProxy(sessionID string, config *ProxyConfig) error
	GetProxy(sessionID string) (*ProxyConfig, error)
}

func NewService(repo Repository, whatsapp WhatsAppManager) *Service {
	return &Service{
		repo:      repo,
		whatsapp:  whatsapp,
		generator: uuid.New(),
	}
}

func (s *Service) CreateSession(ctx context.Context, req *CreateSessionRequest) (*Session, error) {
	// DeviceJid will be set when the session connects to WhatsApp

	// Create new session
	session := NewSession(req.Name)
	session.ProxyConfig = req.ProxyConfig

	// Save to database
	if err := s.repo.Create(ctx, session); err != nil {
		return nil, errors.Wrap(err, "failed to create session")
	}

	// Initialize WhatsApp session
	if err := s.whatsapp.CreateSession(session.ID.String(), req.ProxyConfig); err != nil {
		return nil, errors.Wrap(err, "failed to initialize WhatsApp session")
	}

	return session, nil
}

func (s *Service) GetSession(ctx context.Context, id string) (*SessionInfo, error) {
	session, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get session")
	}

	if session == nil {
		return nil, errors.ErrNotFound
	}

	info := &SessionInfo{
		Session: session,
	}

	if session.IsConnected {
		deviceInfo, _ := s.whatsapp.GetDeviceInfo(id)
		info.DeviceInfo = deviceInfo
	}

	return info, nil
}

func (s *Service) ListSessions(ctx context.Context, req *ListSessionsRequest) ([]*Session, int, error) {
	// Set default values
	if req.Limit == 0 {
		req.Limit = 20
	}

	sessions, total, err := s.repo.List(ctx, req)
	if err != nil {
		return nil, 0, errors.Wrap(err, "failed to list sessions")
	}

	return sessions, total, nil
}

func (s *Service) DeleteSession(ctx context.Context, id string) error {
	session, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return errors.Wrap(err, "failed to get session")
	}

	if session == nil {
		return errors.ErrNotFound
	}

	// Disconnect and cleanup WhatsApp session
	if session.IsActive() {
		if err := s.whatsapp.DisconnectSession(id); err != nil {
			// Log error but continue with deletion
			// We don't want to fail deletion just because disconnect failed
			_ = err // Explicitly ignore error
		}
	}

	// Delete from database
	if err := s.repo.Delete(ctx, id); err != nil {
		return errors.Wrap(err, "failed to delete session")
	}

	return nil
}

func (s *Service) ConnectSession(ctx context.Context, id string) error {
	session, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return errors.Wrap(err, "failed to get session")
	}

	if session == nil {
		return errors.ErrNotFound
	}

	if !session.CanConnect() {
		return errors.NewWithDetails(400, "Cannot connect session", "Session is already connected")
	}

	// Mark as connecting (will be updated to connected after successful connection)

	// Connect to WhatsApp
	if err := s.whatsapp.ConnectSession(id); err != nil {
		session.SetConnectionError(err.Error())
		if updateErr := s.repo.Update(ctx, session); updateErr != nil {
			// Log the update error but return the original connection error
			_ = updateErr // Explicitly ignore update error
		}
		return errors.Wrap(err, "failed to connect to WhatsApp")
	}

	// Update status to connected
	session.SetConnected(true)
	if err := s.repo.Update(ctx, session); err != nil {
		return errors.Wrap(err, "failed to update session status")
	}

	return nil
}

func (s *Service) LogoutSession(ctx context.Context, id string) error {
	session, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return errors.Wrap(err, "failed to get session")
	}

	if session == nil {
		return errors.ErrNotFound
	}

	if !session.CanLogout() {
		return errors.NewWithDetails(400, "Cannot logout session", "Session is not connected")
	}

	// Logout from WhatsApp
	if err := s.whatsapp.LogoutSession(id); err != nil {
		return errors.Wrap(err, "failed to logout from WhatsApp")
	}

	// Update status to disconnected
	session.SetConnected(false)
	if err := s.repo.Update(ctx, session); err != nil {
		return errors.Wrap(err, "failed to update session status")
	}

	return nil
}

func (s *Service) GetQRCode(ctx context.Context, id string) (*QRCodeResponse, error) {
	session, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get session")
	}

	if session == nil {
		return nil, errors.ErrNotFound
	}

	qrResponse, err := s.whatsapp.GetQRCode(id)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get QR code")
	}

	return qrResponse, nil
}

func (s *Service) PairPhone(ctx context.Context, id string, req *PairPhoneRequest) error {
	session, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return errors.Wrap(err, "failed to get session")
	}

	if session == nil {
		return errors.ErrNotFound
	}

	// Pair phone with WhatsApp
	if err := s.whatsapp.PairPhone(id, req.PhoneNumber); err != nil {
		return errors.Wrap(err, "failed to pair phone")
	}

	// Update session with device JID (will be set by WhatsApp manager)
	session.UpdatedAt = time.Now()
	if err := s.repo.Update(ctx, session); err != nil {
		return errors.Wrap(err, "failed to update session")
	}

	return nil
}

func (s *Service) SetProxy(ctx context.Context, id string, config *ProxyConfig) error {
	session, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return errors.Wrap(err, "failed to get session")
	}

	if session == nil {
		return errors.ErrNotFound
	}

	// Set proxy in WhatsApp
	if err := s.whatsapp.SetProxy(id, config); err != nil {
		return errors.Wrap(err, "failed to set proxy")
	}

	// Update session
	session.ProxyConfig = config
	session.UpdatedAt = time.Now()
	if err := s.repo.Update(ctx, session); err != nil {
		return errors.Wrap(err, "failed to update session")
	}

	return nil
}

func (s *Service) GetProxy(ctx context.Context, id string) (*ProxyConfig, error) {
	session, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get session")
	}

	if session == nil {
		return nil, errors.ErrNotFound
	}

	return session.ProxyConfig, nil
}
