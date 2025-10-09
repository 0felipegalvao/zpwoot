package cache

import (
	"context"
	"fmt"
	"time"

	"zpwoot/internal/core/domain/proxy"
	"zpwoot/internal/core/ports/output"
)

type CachedProxyRepository struct {
	repository proxy.Repository
	cache      output.CachePort
	keyBuilder output.CacheKeyBuilder
	serializer output.CacheSerializer
	logger     output.Logger
	config     *CacheRepositoryConfig
}

func NewCachedProxyRepository(
	repository proxy.Repository,
	cache output.CachePort,
	keyBuilder output.CacheKeyBuilder,
	serializer output.CacheSerializer,
	logger output.Logger,
	config *CacheRepositoryConfig,
) proxy.Repository {
	if config == nil {
		config = &CacheRepositoryConfig{
			SessionTTL: 15 * time.Minute,
			ListTTL:    5 * time.Minute,
			Enabled:    true,
		}
	}

	return &CachedProxyRepository{
		repository: repository,
		cache:      cache,
		keyBuilder: keyBuilder,
		serializer: serializer,
		logger:     logger,
		config:     config,
	}
}

func (r *CachedProxyRepository) Create(ctx context.Context, config *proxy.ProxyConfig) error {

	err := r.repository.Create(ctx, config)
	if err != nil {
		return err
	}

	if !r.config.Enabled {
		return nil
	}

	if err := r.cacheProxyConfig(ctx, config); err != nil {
		r.logger.Warn().Err(err).Str("session_id", config.SessionID).Msg("Failed to cache proxy config after create")
	}

	return nil
}

func (r *CachedProxyRepository) GetBySessionID(ctx context.Context, sessionID string) (*proxy.ProxyConfig, error) {
	if !r.config.Enabled {
		return r.repository.GetBySessionID(ctx, sessionID)
	}

	key := r.keyBuilder.ProxyKey(sessionID)

	if config, err := r.getProxyConfigFromCache(ctx, key); err == nil {
		r.logger.Debug().Str("session_id", sessionID).Msg("Proxy config found in cache")
		return config, nil
	}

	config, err := r.repository.GetBySessionID(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	if err := r.cacheProxyConfig(ctx, config); err != nil {
		r.logger.Warn().Err(err).Str("session_id", sessionID).Msg("Failed to cache proxy config")
	}

	r.logger.Debug().Str("session_id", sessionID).Msg("Proxy config loaded from database and cached")
	return config, nil
}

func (r *CachedProxyRepository) Update(ctx context.Context, config *proxy.ProxyConfig) error {

	err := r.repository.Update(ctx, config)
	if err != nil {
		return err
	}

	if !r.config.Enabled {
		return nil
	}

	r.invalidateProxyConfigCache(ctx, config.SessionID)

	if err := r.cacheProxyConfig(ctx, config); err != nil {
		r.logger.Warn().Err(err).Str("session_id", config.SessionID).Msg("Failed to cache proxy config after update")
	}

	return nil
}

func (r *CachedProxyRepository) Delete(ctx context.Context, sessionID string) error {

	err := r.repository.Delete(ctx, sessionID)
	if err != nil {
		return err
	}

	if r.config.Enabled {

		r.invalidateProxyConfigCache(ctx, sessionID)
	}

	return nil
}

func (r *CachedProxyRepository) cacheProxyConfig(ctx context.Context, config *proxy.ProxyConfig) error {
	key := r.keyBuilder.ProxyKey(config.SessionID)

	data, err := r.serializer.Serialize(config)
	if err != nil {
		return fmt.Errorf("failed to serialize proxy config: %w", err)
	}

	return r.cache.Set(ctx, key, data, r.config.SessionTTL)
}

func (r *CachedProxyRepository) getProxyConfigFromCache(ctx context.Context, key string) (*proxy.ProxyConfig, error) {
	data, err := r.cache.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	var config proxy.ProxyConfig
	if err := r.serializer.Deserialize(data, &config); err != nil {
		return nil, fmt.Errorf("failed to deserialize proxy config: %w", err)
	}

	return &config, nil
}

func (r *CachedProxyRepository) invalidateProxyConfigCache(ctx context.Context, sessionID string) {
	key := r.keyBuilder.ProxyKey(sessionID)

	if err := r.cache.Delete(ctx, key); err != nil {
		r.logger.Warn().Err(err).Str("session_id", sessionID).Msg("Failed to invalidate proxy config cache")
	}
}

func (r *CachedProxyRepository) InvalidateAllProxyCache(ctx context.Context) error {
	if !r.config.Enabled {
		return nil
	}

	pattern := r.keyBuilder.ProxyPattern()
	return r.cache.DeleteByPattern(ctx, pattern)
}

func (r *CachedProxyRepository) WarmupCache(ctx context.Context, sessionIDs []string) error {
	if !r.config.Enabled || len(sessionIDs) == 0 {
		return nil
	}

	for _, sessionID := range sessionIDs {

		_, err := r.GetBySessionID(ctx, sessionID)
		if err != nil {
			r.logger.Warn().Err(err).Str("session_id", sessionID).Msg("Failed to warmup proxy config cache")
			continue
		}
	}

	r.logger.Info().Int("count", len(sessionIDs)).Msg("Proxy config cache warmup completed")
	return nil
}

func (r *CachedProxyRepository) GetCacheStats(ctx context.Context) (map[string]interface{}, error) {
	if !r.config.Enabled {
		return map[string]interface{}{
			"enabled": false,
		}, nil
	}

	pattern := r.keyBuilder.ProxyPattern()
	keys, err := r.cache.Keys(ctx, pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to get proxy cache keys: %w", err)
	}

	return map[string]interface{}{
		"enabled":     true,
		"cached_keys": len(keys),
		"pattern":     pattern,
		"ttl":         r.config.SessionTTL.String(),
	}, nil
}
