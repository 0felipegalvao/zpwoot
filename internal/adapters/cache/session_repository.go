package cache

import (
	"context"
	"fmt"
	"time"

	"zpwoot/internal/core/domain/session"
	"zpwoot/internal/core/ports/output"
)

type CachedSessionRepository struct {
	repository session.Repository
	cache      output.CachePort
	keyBuilder output.CacheKeyBuilder
	serializer output.CacheSerializer
	logger     output.Logger
	config     *CacheRepositoryConfig
}

type CacheRepositoryConfig struct {
	SessionTTL time.Duration
	ListTTL    time.Duration
	Enabled    bool
}

func NewCachedSessionRepository(
	repository session.Repository,
	cache output.CachePort,
	keyBuilder output.CacheKeyBuilder,
	serializer output.CacheSerializer,
	logger output.Logger,
	config *CacheRepositoryConfig,
) session.Repository {
	if config == nil {
		config = &CacheRepositoryConfig{
			SessionTTL: 5 * time.Minute,
			ListTTL:    1 * time.Minute,
			Enabled:    true,
		}
	}

	return &CachedSessionRepository{
		repository: repository,
		cache:      cache,
		keyBuilder: keyBuilder,
		serializer: serializer,
		logger:     logger,
		config:     config,
	}
}

func (r *CachedSessionRepository) Create(ctx context.Context, sess *session.Session) error {

	err := r.repository.Create(ctx, sess)
	if err != nil {
		return err
	}

	if !r.config.Enabled {
		return nil
	}

	if err := r.cacheSession(ctx, sess); err != nil {
		r.logger.Warn().Err(err).Str("session_id", sess.ID).Msg("Failed to cache session after create")
	}

	r.invalidateListCache(ctx)

	return nil
}

func (r *CachedSessionRepository) GetByID(ctx context.Context, id string) (*session.Session, error) {
	if !r.config.Enabled {
		return r.repository.GetByID(ctx, id)
	}

	key := r.keyBuilder.SessionKey(id)

	if sess, err := r.getSessionFromCache(ctx, key); err == nil {
		r.logger.Debug().Str("session_id", id).Msg("Session found in cache")
		return sess, nil
	}

	sess, err := r.repository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := r.cacheSession(ctx, sess); err != nil {
		r.logger.Warn().Err(err).Str("session_id", id).Msg("Failed to cache session")
	}

	r.logger.Debug().Str("session_id", id).Msg("Session loaded from database and cached")
	return sess, nil
}

func (r *CachedSessionRepository) GetByName(ctx context.Context, name string) (*session.Session, error) {
	if !r.config.Enabled {
		return r.repository.GetByName(ctx, name)
	}

	key := r.keyBuilder.SessionByNameKey(name)

	if sess, err := r.getSessionFromCache(ctx, key); err == nil {
		r.logger.Debug().Str("session_name", name).Msg("Session found in cache by name")
		return sess, nil
	}

	sess, err := r.repository.GetByName(ctx, name)
	if err != nil {
		return nil, err
	}

	if err := r.cacheSessionWithMultipleKeys(ctx, sess); err != nil {
		r.logger.Warn().Err(err).Str("session_name", name).Msg("Failed to cache session")
	}

	r.logger.Debug().Str("session_name", name).Msg("Session loaded from database and cached")
	return sess, nil
}

func (r *CachedSessionRepository) GetByJID(ctx context.Context, jid string) (*session.Session, error) {
	if !r.config.Enabled {
		return r.repository.GetByJID(ctx, jid)
	}

	key := r.keyBuilder.SessionByJIDKey(jid)

	if sess, err := r.getSessionFromCache(ctx, key); err == nil {
		r.logger.Debug().Str("session_jid", jid).Msg("Session found in cache by JID")
		return sess, nil
	}

	sess, err := r.repository.GetByJID(ctx, jid)
	if err != nil {
		return nil, err
	}

	if err := r.cacheSessionWithMultipleKeys(ctx, sess); err != nil {
		r.logger.Warn().Err(err).Str("session_jid", jid).Msg("Failed to cache session")
	}

	r.logger.Debug().Str("session_jid", jid).Msg("Session loaded from database and cached")
	return sess, nil
}

func (r *CachedSessionRepository) Update(ctx context.Context, sess *session.Session) error {

	err := r.repository.Update(ctx, sess)
	if err != nil {
		return err
	}

	if !r.config.Enabled {
		return nil
	}

	r.invalidateSessionCache(ctx, sess)

	if err := r.cacheSessionWithMultipleKeys(ctx, sess); err != nil {
		r.logger.Warn().Err(err).Str("session_id", sess.ID).Msg("Failed to cache session after update")
	}

	r.invalidateListCache(ctx)

	return nil
}

func (r *CachedSessionRepository) Delete(ctx context.Context, id string) error {

	sess, err := r.GetByID(ctx, id)
	if err == nil && r.config.Enabled {
		r.invalidateSessionCache(ctx, sess)
	}

	err = r.repository.Delete(ctx, id)
	if err != nil {
		return err
	}

	if r.config.Enabled {

		r.invalidateListCache(ctx)
	}

	return nil
}

func (r *CachedSessionRepository) List(ctx context.Context, limit, offset int) ([]*session.Session, error) {
	if !r.config.Enabled {
		return r.repository.List(ctx, limit, offset)
	}

	key := r.keyBuilder.SessionListKey(limit, offset)

	if sessions, err := r.getSessionListFromCache(ctx, key); err == nil {
		r.logger.Debug().Int("limit", limit).Int("offset", offset).Msg("Session list found in cache")
		return sessions, nil
	}

	sessions, err := r.repository.List(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	if err := r.cacheSessionList(ctx, key, sessions); err != nil {
		r.logger.Warn().Err(err).Int("limit", limit).Int("offset", offset).Msg("Failed to cache session list")
	}

	r.logger.Debug().Int("limit", limit).Int("offset", offset).Msg("Session list loaded from database and cached")
	return sessions, nil
}

func (r *CachedSessionRepository) cacheSession(ctx context.Context, sess *session.Session) error {
	key := r.keyBuilder.SessionKey(sess.ID)
	return r.cacheSessionWithKey(ctx, key, sess)
}

func (r *CachedSessionRepository) cacheSessionWithMultipleKeys(ctx context.Context, sess *session.Session) error {
	keys := []string{
		r.keyBuilder.SessionKey(sess.ID),
		r.keyBuilder.SessionByNameKey(sess.Name),
	}

	if sess.DeviceJID != "" {
		keys = append(keys, r.keyBuilder.SessionByJIDKey(sess.DeviceJID))
	}

	data, err := r.serializer.Serialize(sess)
	if err != nil {
		return fmt.Errorf("failed to serialize session: %w", err)
	}

	items := make(map[string][]byte)
	for _, key := range keys {
		items[key] = data
	}

	return r.cache.MSet(ctx, items, r.config.SessionTTL)
}

func (r *CachedSessionRepository) cacheSessionWithKey(ctx context.Context, key string, sess *session.Session) error {
	data, err := r.serializer.Serialize(sess)
	if err != nil {
		return fmt.Errorf("failed to serialize session: %w", err)
	}

	return r.cache.Set(ctx, key, data, r.config.SessionTTL)
}

func (r *CachedSessionRepository) getSessionFromCache(ctx context.Context, key string) (*session.Session, error) {
	data, err := r.cache.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	var sess session.Session
	if err := r.serializer.Deserialize(data, &sess); err != nil {
		return nil, fmt.Errorf("failed to deserialize session: %w", err)
	}

	return &sess, nil
}

func (r *CachedSessionRepository) invalidateSessionCache(ctx context.Context, sess *session.Session) {
	keys := []string{
		r.keyBuilder.SessionKey(sess.ID),
		r.keyBuilder.SessionByNameKey(sess.Name),
	}

	if sess.DeviceJID != "" {
		keys = append(keys, r.keyBuilder.SessionByJIDKey(sess.DeviceJID))
	}

	if err := r.cache.MDelete(ctx, keys); err != nil {
		r.logger.Warn().Err(err).Str("session_id", sess.ID).Msg("Failed to invalidate session cache")
	}
}

func (r *CachedSessionRepository) cacheSessionList(ctx context.Context, key string, sessions []*session.Session) error {
	data, err := r.serializer.Serialize(sessions)
	if err != nil {
		return fmt.Errorf("failed to serialize session list: %w", err)
	}

	return r.cache.Set(ctx, key, data, r.config.ListTTL)
}

func (r *CachedSessionRepository) getSessionListFromCache(ctx context.Context, key string) ([]*session.Session, error) {
	data, err := r.cache.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	var sessions []*session.Session
	if err := r.serializer.Deserialize(data, &sessions); err != nil {
		return nil, fmt.Errorf("failed to deserialize session list: %w", err)
	}

	return sessions, nil
}

func (r *CachedSessionRepository) invalidateListCache(ctx context.Context) {
	pattern := r.keyBuilder.BuildKey("session", "list", "*")
	if err := r.cache.DeleteByPattern(ctx, pattern); err != nil {
		r.logger.Warn().Err(err).Msg("Failed to invalidate session list cache")
	}
}
