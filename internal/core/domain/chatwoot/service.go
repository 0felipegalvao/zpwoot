package chatwoot

import (
	"context"
	"fmt"

	"zpwoot/internal/core/domain/common"
)

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{
		repository: repository,
	}
}

func (s *Service) CreateConfiguration(ctx context.Context, sessionID, url, token, accountID string) (*Chatwoot, error) {

	exists, err := s.repository.Exists(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to check if chatwoot configuration exists: %w", err)
	}

	if exists {
		return nil, common.ErrAlreadyExists
	}

	chatwoot := NewChatwoot(sessionID, url, token, accountID)

	if err := s.repository.Create(ctx, chatwoot); err != nil {
		return nil, fmt.Errorf("failed to create chatwoot configuration: %w", err)
	}

	return chatwoot, nil
}

func (s *Service) GetConfiguration(ctx context.Context, sessionID string) (*Chatwoot, error) {
	chatwoot, err := s.repository.GetBySessionID(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get chatwoot configuration: %w", err)
	}

	return chatwoot, nil
}

func (s *Service) GetConfigurationByID(ctx context.Context, id string) (*Chatwoot, error) {
	chatwoot, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get chatwoot configuration by ID: %w", err)
	}

	return chatwoot, nil
}

func (s *Service) UpdateConfiguration(ctx context.Context, sessionID, url, token, accountID string, inboxID *string) (*Chatwoot, error) {
	chatwoot, err := s.repository.GetBySessionID(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get chatwoot configuration: %w", err)
	}

	chatwoot.UpdateConfiguration(url, token, accountID, inboxID)

	if err := s.repository.Update(ctx, chatwoot); err != nil {
		return nil, fmt.Errorf("failed to update chatwoot configuration: %w", err)
	}

	return chatwoot, nil
}

func (s *Service) UpdateAdvancedSettings(ctx context.Context, sessionID string, settings *AdvancedSettings) (*Chatwoot, error) {
	chatwoot, err := s.repository.GetBySessionID(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get chatwoot configuration: %w", err)
	}

	chatwoot.UpdateAdvancedSettings(settings)

	if err := s.repository.Update(ctx, chatwoot); err != nil {
		return nil, fmt.Errorf("failed to update chatwoot advanced settings: %w", err)
	}

	return chatwoot, nil
}

func (s *Service) EnableConfiguration(ctx context.Context, sessionID string) (*Chatwoot, error) {
	chatwoot, err := s.repository.GetBySessionID(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get chatwoot configuration: %w", err)
	}

	chatwoot.Enable()

	if err := s.repository.Update(ctx, chatwoot); err != nil {
		return nil, fmt.Errorf("failed to enable chatwoot configuration: %w", err)
	}

	return chatwoot, nil
}

func (s *Service) DisableConfiguration(ctx context.Context, sessionID string) (*Chatwoot, error) {
	chatwoot, err := s.repository.GetBySessionID(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get chatwoot configuration: %w", err)
	}

	chatwoot.Disable()

	if err := s.repository.Update(ctx, chatwoot); err != nil {
		return nil, fmt.Errorf("failed to disable chatwoot configuration: %w", err)
	}

	return chatwoot, nil
}

func (s *Service) DeleteConfiguration(ctx context.Context, sessionID string) error {
	if err := s.repository.DeleteBySessionID(ctx, sessionID); err != nil {
		return fmt.Errorf("failed to delete chatwoot configuration: %w", err)
	}

	return nil
}

func (s *Service) ListConfigurations(ctx context.Context, limit, offset int) ([]*Chatwoot, error) {
	configurations, err := s.repository.List(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list chatwoot configurations: %w", err)
	}

	return configurations, nil
}

func (s *Service) ListEnabledConfigurations(ctx context.Context, limit, offset int) ([]*Chatwoot, error) {
	configurations, err := s.repository.ListByEnabled(ctx, true, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list enabled chatwoot configurations: %w", err)
	}

	return configurations, nil
}

func (s *Service) ConfigurationExists(ctx context.Context, sessionID string) (bool, error) {
	exists, err := s.repository.Exists(ctx, sessionID)
	if err != nil {
		return false, fmt.Errorf("failed to check if chatwoot configuration exists: %w", err)
	}

	return exists, nil
}
