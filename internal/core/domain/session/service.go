package session

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"zpwoot/internal/core/domain/common"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, name string) (*Session, error) {
	if name == "" {
		return nil, common.ErrInvalidInput
	}

	if err := s.checkNameAvailable(ctx, name); err != nil {
		return nil, err
	}

	session := NewSession(name)
	if err := s.repo.Create(ctx, session); err != nil {
		if isUniqueConstraintError(err) {
			return nil, common.ErrAlreadyExists
		}
		return nil, fmt.Errorf("create session: %w", err)
	}

	return session, nil
}

func (s *Service) CreateFromSession(ctx context.Context, session *Session) error {
	if session.Name == "" {
		return common.ErrInvalidInput
	}

	if err := s.checkNameAvailable(ctx, session.Name); err != nil {
		return err
	}

	if err := s.repo.Create(ctx, session); err != nil {
		if isUniqueConstraintError(err) {
			return common.ErrAlreadyExists
		}
		return fmt.Errorf("create session: %w", err)
	}

	return nil
}

func (s *Service) Get(ctx context.Context, id string) (*Session, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, session *Session) error {
	return s.repo.Update(ctx, session)
}

func (s *Service) List(ctx context.Context, limit, offset int) ([]*Session, error) {
	return s.repo.List(ctx, limit, offset)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *Service) checkNameAvailable(ctx context.Context, name string) error {
	existing, err := s.repo.GetByName(ctx, name)
	if err != nil && !errors.Is(err, common.ErrNotFound) {
		return fmt.Errorf("check session name: %w", err)
	}
	if existing != nil {
		return common.ErrAlreadyExists
	}
	return nil
}

func isUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "duplicate key") ||
		strings.Contains(errStr, "unique constraint") ||
		strings.Contains(errStr, "violates unique")
}
