package cache

import (
	"context"
	"fmt"
	"time"

	"zpwoot/internal/core/domain/chatwoot"
	"zpwoot/internal/core/ports/output"
)

type CachedChatwootRepository struct {
	repository chatwoot.Repository
	cache      output.CachePort
	keyBuilder output.CacheKeyBuilder
	serializer output.CacheSerializer
	logger     output.Logger
	config     *CacheRepositoryConfig
}

func NewCachedChatwootRepository(
	repository chatwoot.Repository,
	cache output.CachePort,
	keyBuilder output.CacheKeyBuilder,
	serializer output.CacheSerializer,
	logger output.Logger,
	config *CacheRepositoryConfig,
) chatwoot.Repository {
	if config == nil {
		config = &CacheRepositoryConfig{
			SessionTTL: 15 * time.Minute,
			ListTTL:    5 * time.Minute,
			Enabled:    true,
		}
	}

	return &CachedChatwootRepository{
		repository: repository,
		cache:      cache,
		keyBuilder: keyBuilder,
		serializer: serializer,
		logger:     logger,
		config:     config,
	}
}

func (r *CachedChatwootRepository) Create(ctx context.Context, c *chatwoot.Chatwoot) error {

	err := r.repository.Create(ctx, c)
	if err != nil {
		return err
	}

	if !r.config.Enabled {
		return nil
	}

	if err := r.cacheChatwootWithMultipleKeys(ctx, c); err != nil {
		r.logger.Warn().Err(err).Str("chatwoot_id", c.ID).Msg("Failed to cache chatwoot after create")
	}

	r.invalidateListCache(ctx)

	return nil
}

func (r *CachedChatwootRepository) GetByID(ctx context.Context, id string) (*chatwoot.Chatwoot, error) {
	if !r.config.Enabled {
		return r.repository.GetByID(ctx, id)
	}

	key := r.keyBuilder.ChatwootKey(id)

	if c, err := r.getChatwootFromCache(ctx, key); err == nil {
		r.logger.Debug().Str("chatwoot_id", id).Msg("Chatwoot found in cache")
		return c, nil
	}

	c, err := r.repository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := r.cacheChatwootWithMultipleKeys(ctx, c); err != nil {
		r.logger.Warn().Err(err).Str("chatwoot_id", id).Msg("Failed to cache chatwoot")
	}

	r.logger.Debug().Str("chatwoot_id", id).Msg("Chatwoot loaded from database and cached")
	return c, nil
}

func (r *CachedChatwootRepository) GetBySessionID(ctx context.Context, sessionID string) (*chatwoot.Chatwoot, error) {
	if !r.config.Enabled {
		return r.repository.GetBySessionID(ctx, sessionID)
	}

	key := r.keyBuilder.ChatwootBySessionKey(sessionID)

	if c, err := r.getChatwootFromCache(ctx, key); err == nil {
		r.logger.Debug().Str("session_id", sessionID).Msg("Chatwoot found in cache by session ID")
		return c, nil
	}

	c, err := r.repository.GetBySessionID(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	if err := r.cacheChatwootWithMultipleKeys(ctx, c); err != nil {
		r.logger.Warn().Err(err).Str("session_id", sessionID).Msg("Failed to cache chatwoot")
	}

	r.logger.Debug().Str("session_id", sessionID).Msg("Chatwoot loaded from database and cached by session ID")
	return c, nil
}

func (r *CachedChatwootRepository) Update(ctx context.Context, c *chatwoot.Chatwoot) error {

	err := r.repository.Update(ctx, c)
	if err != nil {
		return err
	}

	if !r.config.Enabled {
		return nil
	}

	if err := r.cacheChatwootWithMultipleKeys(ctx, c); err != nil {
		r.logger.Warn().Err(err).Str("chatwoot_id", c.ID).Msg("Failed to update chatwoot cache")
	}

	r.invalidateListCache(ctx)

	return nil
}

func (r *CachedChatwootRepository) Delete(ctx context.Context, id string) error {

	c, err := r.GetByID(ctx, id)
	if err == nil && r.config.Enabled {
		r.invalidateChatwootCache(ctx, c)
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

func (r *CachedChatwootRepository) DeleteBySessionID(ctx context.Context, sessionID string) error {

	c, err := r.GetBySessionID(ctx, sessionID)
	if err == nil && r.config.Enabled {
		r.invalidateChatwootCache(ctx, c)
	}

	err = r.repository.DeleteBySessionID(ctx, sessionID)
	if err != nil {
		return err
	}

	if r.config.Enabled {

		r.invalidateListCache(ctx)
	}

	return nil
}

func (r *CachedChatwootRepository) List(ctx context.Context, limit, offset int) ([]*chatwoot.Chatwoot, error) {
	if !r.config.Enabled {
		return r.repository.List(ctx, limit, offset)
	}

	key := r.keyBuilder.ChatwootListKey(limit, offset)

	if configurations, err := r.getChatwootListFromCache(ctx, key); err == nil {
		r.logger.Debug().Int("limit", limit).Int("offset", offset).Msg("Chatwoot list found in cache")
		return configurations, nil
	}

	configurations, err := r.repository.List(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	if err := r.cacheChatwootList(ctx, key, configurations); err != nil {
		r.logger.Warn().Err(err).Msg("Failed to cache chatwoot list")
	}

	r.logger.Debug().Int("limit", limit).Int("offset", offset).Msg("Chatwoot list loaded from database and cached")
	return configurations, nil
}

func (r *CachedChatwootRepository) ListByEnabled(ctx context.Context, enabled bool, limit, offset int) ([]*chatwoot.Chatwoot, error) {
	if !r.config.Enabled {
		return r.repository.ListByEnabled(ctx, enabled, limit, offset)
	}

	key := r.keyBuilder.ChatwootListByEnabledKey(enabled, limit, offset)

	if configurations, err := r.getChatwootListFromCache(ctx, key); err == nil {
		r.logger.Debug().Bool("enabled", enabled).Int("limit", limit).Int("offset", offset).Msg("Chatwoot list by enabled found in cache")
		return configurations, nil
	}

	configurations, err := r.repository.ListByEnabled(ctx, enabled, limit, offset)
	if err != nil {
		return nil, err
	}

	if err := r.cacheChatwootList(ctx, key, configurations); err != nil {
		r.logger.Warn().Err(err).Msg("Failed to cache chatwoot list by enabled")
	}

	r.logger.Debug().Bool("enabled", enabled).Int("limit", limit).Int("offset", offset).Msg("Chatwoot list by enabled loaded from database and cached")
	return configurations, nil
}

func (r *CachedChatwootRepository) Exists(ctx context.Context, sessionID string) (bool, error) {
	if !r.config.Enabled {
		return r.repository.Exists(ctx, sessionID)
	}

	key := r.keyBuilder.ChatwootBySessionKey(sessionID)
	if _, err := r.getChatwootFromCache(ctx, key); err == nil {
		return true, nil
	}

	return r.repository.Exists(ctx, sessionID)
}

func (r *CachedChatwootRepository) cacheChatwootWithMultipleKeys(ctx context.Context, c *chatwoot.Chatwoot) error {
	data, err := r.serializer.Serialize(c)
	if err != nil {
		return fmt.Errorf("failed to serialize chatwoot: %w", err)
	}

	keys := []string{
		r.keyBuilder.ChatwootKey(c.ID),
		r.keyBuilder.ChatwootBySessionKey(c.SessionID),
	}

	for _, key := range keys {
		if err := r.cache.Set(ctx, key, data, r.config.SessionTTL); err != nil {
			r.logger.Warn().Err(err).Str("key", key).Msg("Failed to cache chatwoot with key")
		}
	}

	return nil
}

func (r *CachedChatwootRepository) getChatwootFromCache(ctx context.Context, key string) (*chatwoot.Chatwoot, error) {
	data, err := r.cache.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	var c chatwoot.Chatwoot
	if err := r.serializer.Deserialize(data, &c); err != nil {
		return nil, fmt.Errorf("failed to deserialize chatwoot: %w", err)
	}

	return &c, nil
}

func (r *CachedChatwootRepository) invalidateChatwootCache(ctx context.Context, c *chatwoot.Chatwoot) {
	keys := []string{
		r.keyBuilder.ChatwootKey(c.ID),
		r.keyBuilder.ChatwootBySessionKey(c.SessionID),
	}

	if err := r.cache.MDelete(ctx, keys); err != nil {
		r.logger.Warn().Err(err).Str("chatwoot_id", c.ID).Msg("Failed to invalidate chatwoot cache")
	}
}

func (r *CachedChatwootRepository) getChatwootListFromCache(ctx context.Context, key string) ([]*chatwoot.Chatwoot, error) {
	data, err := r.cache.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	var configurations []*chatwoot.Chatwoot
	if err := r.serializer.Deserialize(data, &configurations); err != nil {
		return nil, fmt.Errorf("failed to deserialize chatwoot list: %w", err)
	}

	return configurations, nil
}

func (r *CachedChatwootRepository) invalidateListCache(ctx context.Context) {
	pattern := r.keyBuilder.BuildKey("chatwoot", "list", "*")
	if err := r.cache.DeleteByPattern(ctx, pattern); err != nil {
		r.logger.Warn().Err(err).Msg("Failed to invalidate chatwoot list cache")
	}
}

func (r *CachedChatwootRepository) cacheChatwootList(ctx context.Context, key string, configurations []*chatwoot.Chatwoot) error {
	data, err := r.serializer.Serialize(configurations)
	if err != nil {
		return fmt.Errorf("failed to serialize chatwoot list: %w", err)
	}

	return r.cache.Set(ctx, key, data, r.config.ListTTL)
}
