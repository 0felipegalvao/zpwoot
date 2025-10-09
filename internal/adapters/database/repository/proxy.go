package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"zpwoot/internal/core/domain/common"
	"zpwoot/internal/core/domain/proxy"

	"github.com/jmoiron/sqlx"
)

type ProxyRepository struct {
	db *sqlx.DB
}

func NewProxyRepository(db *sqlx.DB) *ProxyRepository {
	return &ProxyRepository{db: db}
}

func (r *ProxyRepository) Create(ctx context.Context, config *proxy.ProxyConfig) error {
	query := `
		INSERT INTO "zpProxyConfig" (
			"id", "sessionId", "host", "port", "protocol",
			"username", "password", "enabled", "createdAt", "updatedAt"
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		)
	`

	_, err := r.db.ExecContext(ctx, query,
		config.ID,
		config.SessionID,
		config.Host,
		config.Port,
		config.Protocol,
		config.Username,
		config.Password,
		config.Enabled,
		config.CreatedAt,
		config.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create proxy config: %w", err)
	}

	return nil
}

func (r *ProxyRepository) GetBySessionID(ctx context.Context, sessionID string) (*proxy.ProxyConfig, error) {
	query := `
		SELECT "id", "sessionId", "host", "port", "protocol",
			   "username", "password", "enabled", "createdAt", "updatedAt"
		FROM "zpProxyConfig"
		WHERE "sessionId" = $1
	`

	var config proxy.ProxyConfig

	err := r.db.GetContext(ctx, &config, query, sessionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, common.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get proxy config: %w", err)
	}

	return &config, nil
}

func (r *ProxyRepository) Update(ctx context.Context, config *proxy.ProxyConfig) error {
	query := `
		UPDATE "zpProxyConfig" SET
			"host" = $2,
			"port" = $3,
			"protocol" = $4,
			"username" = $5,
			"password" = $6,
			"enabled" = $7,
			"updatedAt" = $8
		WHERE "sessionId" = $1
	`

	result, err := r.db.ExecContext(ctx, query,
		config.SessionID,
		config.Host,
		config.Port,
		config.Protocol,
		config.Username,
		config.Password,
		config.Enabled,
		config.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update proxy config: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return common.ErrNotFound
	}

	return nil
}

func (r *ProxyRepository) Delete(ctx context.Context, sessionID string) error {
	query := `DELETE FROM "zpProxyConfig" WHERE "sessionId" = $1`

	result, err := r.db.ExecContext(ctx, query, sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete proxy config: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return common.ErrNotFound
	}

	return nil
}
