package proxy

import (
	"context"

	"zpwoot/internal/core/domain/common"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, config *ProxyConfig) error {
	if err := s.validate(config); err != nil {
		return err
	}
	return s.repo.Create(ctx, config)
}

func (s *Service) GetBySessionID(ctx context.Context, sessionID string) (*ProxyConfig, error) {
	return s.repo.GetBySessionID(ctx, sessionID)
}

func (s *Service) Update(ctx context.Context, config *ProxyConfig) error {
	if err := s.validate(config); err != nil {
		return err
	}
	return s.repo.Update(ctx, config)
}

func (s *Service) Delete(ctx context.Context, sessionID string) error {
	return s.repo.Delete(ctx, sessionID)
}

func (s *Service) validate(config *ProxyConfig) error {
	if config.Host == "" {
		return ErrInvalidHost
	}
	if config.Port <= 0 || config.Port > 65535 {
		return ErrInvalidPort
	}
	if !config.Protocol.IsValid() {
		return ErrInvalidProtocol
	}
	if config.SessionID == "" {
		return common.ErrInvalidInput
	}
	return nil
}
