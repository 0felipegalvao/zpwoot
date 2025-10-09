package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"zpwoot/internal/core/domain/common"
	"zpwoot/internal/core/domain/session"

	"github.com/jmoiron/sqlx"
)

type SessionRepository struct {
	db *sqlx.DB
}

func NewSessionRepository(db *sqlx.DB) *SessionRepository {
	return &SessionRepository{
		db: db,
	}
}

func (r *SessionRepository) Create(ctx context.Context, sess *session.Session) error {
	query := `
		INSERT INTO "zpSessions" (
			"id", "name", "apiKey", "deviceJid", "isConnected",
			"connectionError", "qrCode", "qrCodeExpiresAt",
			"createdAt", "updatedAt"
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		)
	`

	_, err := r.db.ExecContext(ctx, query,
		sess.ID,
		sess.Name,
		sess.APIKey,
		sess.DeviceJID,
		sess.IsConnected,
		sess.ConnectionError,
		sess.QRCode,
		sess.QRCodeExpiresAt,
		sess.CreatedAt,
		sess.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	return nil
}

func (r *SessionRepository) GetByID(ctx context.Context, id string) (*session.Session, error) {
	query := `
		SELECT "id", "name", "apiKey", "deviceJid", "isConnected",
			   "connectionError", "qrCode", "qrCodeExpiresAt",
			   "createdAt", "updatedAt"
		FROM "zpSessions"
		WHERE "id" = $1
	`

	var sess session.Session

	err := r.db.GetContext(ctx, &sess, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, common.ErrNotFound
		}

		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return &sess, nil
}

func (r *SessionRepository) GetByJID(ctx context.Context, jid string) (*session.Session, error) {
	query := `
		SELECT "id", "name", "apiKey", "deviceJid", "isConnected",
			   "connectionError", "qrCode", "qrCodeExpiresAt",
			   "createdAt", "updatedAt"
		FROM "zpSessions"
		WHERE "deviceJid" = $1
	`

	var sess session.Session

	err := r.db.GetContext(ctx, &sess, query, jid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, common.ErrNotFound
		}

		return nil, fmt.Errorf("failed to get session by JID: %w", err)
	}

	return &sess, nil
}

func (r *SessionRepository) GetByName(ctx context.Context, name string) (*session.Session, error) {
	query := `
		SELECT "id", "name", "apiKey", "deviceJid", "isConnected",
			   "connectionError", "qrCode", "qrCodeExpiresAt",
			   "createdAt", "updatedAt"
		FROM "zpSessions"
		WHERE "name" = $1
	`

	var sess session.Session

	err := r.db.GetContext(ctx, &sess, query, name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, common.ErrNotFound
		}

		return nil, fmt.Errorf("failed to get session by name: %w", err)
	}

	return &sess, nil
}

func (r *SessionRepository) List(ctx context.Context, limit, offset int) ([]*session.Session, error) {
	query := `
		SELECT "id", "name", "apiKey", "deviceJid", "isConnected",
			   "connectionError", "qrCode", "qrCodeExpiresAt",
			   "createdAt", "updatedAt"
		FROM "zpSessions"
		ORDER BY "createdAt" DESC
		LIMIT $1 OFFSET $2
	`

	var sessions []*session.Session

	err := r.db.SelectContext(ctx, &sessions, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}

	return sessions, nil
}

func (r *SessionRepository) Update(ctx context.Context, sess *session.Session) error {
	query := `
		UPDATE "zpSessions" SET
			"name" = $2,
			"deviceJid" = $3,
			"isConnected" = $4,
			"connectionError" = $5,
			"qrCode" = $6,
			"qrCodeExpiresAt" = $7,
			"updatedAt" = $8
		WHERE "id" = $1
	`

	result, err := r.db.ExecContext(ctx, query,
		sess.ID,
		sess.Name,
		sess.DeviceJID,
		sess.IsConnected,
		sess.ConnectionError,
		sess.QRCode,
		sess.QRCodeExpiresAt,
		time.Now(),
	)

	if err != nil {
		return fmt.Errorf("failed to update session: %w", err)
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

func (r *SessionRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM "zpSessions" WHERE "id" = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
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
