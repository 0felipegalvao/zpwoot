package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"zpwoot/internal/core/domain/chatwoot"
	"zpwoot/internal/core/domain/common"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type ChatwootRepository struct {
	db *sqlx.DB
}

func NewChatwootRepository(db *sqlx.DB) *ChatwootRepository {
	return &ChatwootRepository{
		db: db,
	}
}

type chatwootDB struct {
	ID             string         `db:"id"`
	SessionID      string         `db:"sessionId"`
	URL            string         `db:"url"`
	Token          string         `db:"token"`
	AccountID      string         `db:"accountId"`
	InboxID        sql.NullString `db:"inboxId"`
	Enabled        bool           `db:"enabled"`
	InboxName      sql.NullString `db:"inboxName"`
	AutoCreate     bool           `db:"autoCreate"`
	SignMsg        bool           `db:"signMsg"`
	SignDelimiter  string         `db:"signDelimiter"`
	ReopenConv     bool           `db:"reopenConv"`
	ConvPending    bool           `db:"convPending"`
	ImportContacts bool           `db:"importContacts"`
	ImportMessages bool           `db:"importMessages"`
	ImportDays     int            `db:"importDays"`
	MergeBrazil    bool           `db:"mergeBrazil"`
	Organization   sql.NullString `db:"organization"`
	Logo           sql.NullString `db:"logo"`
	Number         sql.NullString `db:"number"`
	IgnoreJids     pq.StringArray `db:"ignoreJids"`
	CreatedAt      sql.NullTime   `db:"createdAt"`
	UpdatedAt      sql.NullTime   `db:"updatedAt"`
}

func (c *chatwootDB) toDomain() (*chatwoot.Chatwoot, error) {

	ignoreJids := []string(c.IgnoreJids)

	if ignoreJids == nil {
		ignoreJids = []string{}
	}

	entity := &chatwoot.Chatwoot{
		ID:             c.ID,
		SessionID:      c.SessionID,
		URL:            c.URL,
		Token:          c.Token,
		AccountID:      c.AccountID,
		Enabled:        c.Enabled,
		AutoCreate:     c.AutoCreate,
		SignMsg:        c.SignMsg,
		SignDelimiter:  c.SignDelimiter,
		ReopenConv:     c.ReopenConv,
		ConvPending:    c.ConvPending,
		ImportContacts: c.ImportContacts,
		ImportMessages: c.ImportMessages,
		ImportDays:     c.ImportDays,
		MergeBrazil:    c.MergeBrazil,
		IgnoreJids:     ignoreJids,
	}

	if c.InboxID.Valid {
		entity.InboxID = &c.InboxID.String
	}
	if c.InboxName.Valid {
		entity.InboxName = &c.InboxName.String
	}
	if c.Organization.Valid {
		entity.Organization = &c.Organization.String
	}
	if c.Logo.Valid {
		entity.Logo = &c.Logo.String
	}
	if c.Number.Valid {
		entity.Number = &c.Number.String
	}
	if c.CreatedAt.Valid {
		entity.CreatedAt = c.CreatedAt.Time
	}
	if c.UpdatedAt.Valid {
		entity.UpdatedAt = c.UpdatedAt.Time
	}

	return entity, nil
}

func fromDomain(c *chatwoot.Chatwoot) (*chatwootDB, error) {

	var ignoreJids pq.StringArray
	if c.IgnoreJids != nil {
		ignoreJids = pq.StringArray(c.IgnoreJids)
	} else {
		ignoreJids = pq.StringArray{}
	}

	db := &chatwootDB{
		ID:             c.ID,
		SessionID:      c.SessionID,
		URL:            c.URL,
		Token:          c.Token,
		AccountID:      c.AccountID,
		Enabled:        c.Enabled,
		AutoCreate:     c.AutoCreate,
		SignMsg:        c.SignMsg,
		SignDelimiter:  c.SignDelimiter,
		ReopenConv:     c.ReopenConv,
		ConvPending:    c.ConvPending,
		ImportContacts: c.ImportContacts,
		ImportMessages: c.ImportMessages,
		ImportDays:     c.ImportDays,
		MergeBrazil:    c.MergeBrazil,
		IgnoreJids:     ignoreJids,
		CreatedAt:      sql.NullTime{Time: c.CreatedAt, Valid: true},
		UpdatedAt:      sql.NullTime{Time: c.UpdatedAt, Valid: true},
	}

	if c.InboxID != nil {
		db.InboxID = sql.NullString{String: *c.InboxID, Valid: true}
	}
	if c.InboxName != nil {
		db.InboxName = sql.NullString{String: *c.InboxName, Valid: true}
	}
	if c.Organization != nil {
		db.Organization = sql.NullString{String: *c.Organization, Valid: true}
	}
	if c.Logo != nil {
		db.Logo = sql.NullString{String: *c.Logo, Valid: true}
	}
	if c.Number != nil {
		db.Number = sql.NullString{String: *c.Number, Valid: true}
	}

	return db, nil
}

func (r *ChatwootRepository) Create(ctx context.Context, c *chatwoot.Chatwoot) error {
	db, err := fromDomain(c)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO "zpChatwoot" (
			"id", "sessionId", "url", "token", "accountId", "inboxId", "enabled",
			"inboxName", "autoCreate", "signMsg", "signDelimiter", "reopenConv",
			"convPending", "importContacts", "importMessages", "importDays",
			"mergeBrazil", "organization", "logo", "number", "ignoreJids",
			"createdAt", "updatedAt"
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16,
			$17, $18, $19, $20, $21, $22, $23
		)
	`

	_, err = r.db.ExecContext(ctx, query,
		db.ID, db.SessionID, db.URL, db.Token, db.AccountID, db.InboxID, db.Enabled,
		db.InboxName, db.AutoCreate, db.SignMsg, db.SignDelimiter, db.ReopenConv,
		db.ConvPending, db.ImportContacts, db.ImportMessages, db.ImportDays,
		db.MergeBrazil, db.Organization, db.Logo, db.Number, db.IgnoreJids,
		db.CreatedAt, db.UpdatedAt,
	)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == "23505" {
				return common.ErrAlreadyExists
			}
		}
		return fmt.Errorf("failed to create chatwoot configuration: %w", err)
	}

	return nil
}

func (r *ChatwootRepository) GetByID(ctx context.Context, id string) (*chatwoot.Chatwoot, error) {
	query := `
		SELECT "id", "sessionId", "url", "token", "accountId", "inboxId", "enabled",
		       "inboxName", "autoCreate", "signMsg", "signDelimiter", "reopenConv",
		       "convPending", "importContacts", "importMessages", "importDays",
		       "mergeBrazil", "organization", "logo", "number", "ignoreJids",
		       "createdAt", "updatedAt"
		FROM "zpChatwoot"
		WHERE "id" = $1
	`

	var db chatwootDB
	err := r.db.GetContext(ctx, &db, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, common.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get chatwoot configuration: %w", err)
	}

	return db.toDomain()
}

func (r *ChatwootRepository) GetBySessionID(ctx context.Context, sessionID string) (*chatwoot.Chatwoot, error) {
	query := `
		SELECT "id", "sessionId", "url", "token", "accountId", "inboxId", "enabled",
		       "inboxName", "autoCreate", "signMsg", "signDelimiter", "reopenConv",
		       "convPending", "importContacts", "importMessages", "importDays",
		       "mergeBrazil", "organization", "logo", "number", "ignoreJids",
		       "createdAt", "updatedAt"
		FROM "zpChatwoot"
		WHERE "sessionId" = $1
	`

	var db chatwootDB
	err := r.db.GetContext(ctx, &db, query, sessionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, common.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get chatwoot configuration by session ID: %w", err)
	}

	return db.toDomain()
}

func (r *ChatwootRepository) Update(ctx context.Context, c *chatwoot.Chatwoot) error {
	db, err := fromDomain(c)
	if err != nil {
		return err
	}

	query := `
		UPDATE "zpChatwoot" SET
			"url" = $2, "token" = $3, "accountId" = $4, "inboxId" = $5, "enabled" = $6,
			"inboxName" = $7, "autoCreate" = $8, "signMsg" = $9, "signDelimiter" = $10,
			"reopenConv" = $11, "convPending" = $12, "importContacts" = $13,
			"importMessages" = $14, "importDays" = $15, "mergeBrazil" = $16,
			"organization" = $17, "logo" = $18, "number" = $19, "ignoreJids" = $20,
			"updatedAt" = $21
		WHERE "id" = $1
	`

	result, err := r.db.ExecContext(ctx, query,
		db.ID, db.URL, db.Token, db.AccountID, db.InboxID, db.Enabled,
		db.InboxName, db.AutoCreate, db.SignMsg, db.SignDelimiter,
		db.ReopenConv, db.ConvPending, db.ImportContacts, db.ImportMessages,
		db.ImportDays, db.MergeBrazil, db.Organization, db.Logo, db.Number,
		db.IgnoreJids, db.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update chatwoot configuration: %w", err)
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

func (r *ChatwootRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM "zpChatwoot" WHERE "id" = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete chatwoot configuration: %w", err)
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

func (r *ChatwootRepository) DeleteBySessionID(ctx context.Context, sessionID string) error {
	query := `DELETE FROM "zpChatwoot" WHERE "sessionId" = $1`

	result, err := r.db.ExecContext(ctx, query, sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete chatwoot configuration by session ID: %w", err)
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

func (r *ChatwootRepository) List(ctx context.Context, limit, offset int) ([]*chatwoot.Chatwoot, error) {
	query := `
		SELECT "id", "sessionId", "url", "token", "accountId", "inboxId", "enabled",
		       "inboxName", "autoCreate", "signMsg", "signDelimiter", "reopenConv",
		       "convPending", "importContacts", "importMessages", "importDays",
		       "mergeBrazil", "organization", "logo", "number", "ignoreJids",
		       "createdAt", "updatedAt"
		FROM "zpChatwoot"
		ORDER BY "createdAt" DESC
		LIMIT $1 OFFSET $2
	`

	var dbList []chatwootDB
	err := r.db.SelectContext(ctx, &dbList, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list chatwoot configurations: %w", err)
	}

	configurations := make([]*chatwoot.Chatwoot, 0, len(dbList))
	for _, db := range dbList {
		config, err := db.toDomain()
		if err != nil {
			return nil, err
		}
		configurations = append(configurations, config)
	}

	return configurations, nil
}

func (r *ChatwootRepository) ListByEnabled(ctx context.Context, enabled bool, limit, offset int) ([]*chatwoot.Chatwoot, error) {
	query := `
		SELECT "id", "sessionId", "url", "token", "accountId", "inboxId", "enabled",
		       "inboxName", "autoCreate", "signMsg", "signDelimiter", "reopenConv",
		       "convPending", "importContacts", "importMessages", "importDays",
		       "mergeBrazil", "organization", "logo", "number", "ignoreJids",
		       "createdAt", "updatedAt"
		FROM "zpChatwoot"
		WHERE "enabled" = $1
		ORDER BY "createdAt" DESC
		LIMIT $2 OFFSET $3
	`

	var dbList []chatwootDB
	err := r.db.SelectContext(ctx, &dbList, query, enabled, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list chatwoot configurations by enabled status: %w", err)
	}

	configurations := make([]*chatwoot.Chatwoot, 0, len(dbList))
	for _, db := range dbList {
		config, err := db.toDomain()
		if err != nil {
			return nil, err
		}
		configurations = append(configurations, config)
	}

	return configurations, nil
}

func (r *ChatwootRepository) Exists(ctx context.Context, sessionID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM "zpChatwoot" WHERE "sessionId" = $1)`

	var exists bool
	err := r.db.GetContext(ctx, &exists, query, sessionID)
	if err != nil {
		return false, fmt.Errorf("failed to check if chatwoot configuration exists: %w", err)
	}

	return exists, nil
}
