package cache

import (
	"fmt"
	"time"

	"zpwoot/internal/config"
	"zpwoot/internal/core/ports/output"
)

type CacheManager struct {
	config *config.Config
	logger output.Logger
}

func NewCacheManager(config *config.Config, logger output.Logger) *CacheManager {
	return &CacheManager{
		config: config,
		logger: logger,
	}
}

func (m *CacheManager) CreateCache() (output.CachePort, error) {

	if m.config.Cache.RedisEnabled {
		m.logger.Info().Msg("Using Redis cache")
		return m.createRedisCache()
	} else {
		m.logger.Info().Msg("Using in-memory cache")
		return m.createMemoryCache()
	}
}

func (m *CacheManager) CreateKeyBuilder() output.CacheKeyBuilder {
	return NewDefaultKeyBuilder(m.config.Cache.KeyPrefix)
}

func (m *CacheManager) CreateSerializer() output.CacheSerializer {
	return NewJSONSerializer()
}

func (m *CacheManager) CreateCacheConfig() *CacheRepositoryConfig {
	return &CacheRepositoryConfig{
		SessionTTL: m.config.Cache.SessionTTL,
		ListTTL:    m.config.Cache.ListTTL,
		Enabled:    true,
	}
}

func (m *CacheManager) createRedisCache() (output.CachePort, error) {
	cacheConfig := &output.CacheConfig{
		URL:               m.config.Cache.URL,
		Password:          m.config.Cache.Password,
		DB:                m.config.Cache.DB,
		DialTimeout:       m.config.Cache.DialTimeout,
		ReadTimeout:       m.config.Cache.ReadTimeout,
		WriteTimeout:      m.config.Cache.WriteTimeout,
		PoolSize:          m.config.Cache.PoolSize,
		MinIdleConns:      m.config.Cache.MinIdleConns,
		MaxIdleConns:      m.config.Cache.MaxIdleConns,
		DefaultTTL:        m.config.Cache.DefaultTTL,
		SessionTTL:        m.config.Cache.SessionTTL,
		WebhookTTL:        m.config.Cache.WebhookTTL,
		ProxyTTL:          m.config.Cache.ProxyTTL,
		ListTTL:           m.config.Cache.ListTTL,
		KeyPrefix:         m.config.Cache.KeyPrefix,
		EnableCompression: m.config.Cache.EnableCompression,
		EnableMetrics:     m.config.Cache.EnableMetrics,
	}

	cache, err := NewRedisCache(cacheConfig, m.logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create Redis cache: %w", err)
	}

	m.logger.Info().Str("url", m.config.Cache.URL).Msg("Redis cache created successfully")
	return cache, nil
}

func (m *CacheManager) createMemoryCache() (output.CachePort, error) {
	cache := NewMemoryCache(m.logger, &MemoryCacheConfig{
		DefaultTTL:        m.config.Cache.DefaultTTL,
		CleanupInterval:   1 * time.Minute,
		MaxSize:           10000,
		EnableCompression: m.config.Cache.EnableCompression,
		EnableMetrics:     m.config.Cache.EnableMetrics,
	})

	m.logger.Info().Msg("Memory cache created successfully")
	return cache, nil
}

func (m *CacheManager) ValidateConfig() error {

	if m.config.Cache.RedisEnabled {
		if m.config.Cache.URL == "" {
			return fmt.Errorf("Redis URL is required when REDIS_ENABLED=true")
		}
	}

	return nil
}

func (m *CacheManager) GetCacheInfo() map[string]interface{} {
	info := map[string]interface{}{
		"enabled":       true,
		"redis_enabled": m.config.Cache.RedisEnabled,
		"type":          m.config.Cache.Type,
		"key_prefix":    m.config.Cache.KeyPrefix,
	}

	if m.config.Cache.RedisEnabled {
		info["redis_url"] = m.config.Cache.URL
		info["redis_db"] = m.config.Cache.DB
	}

	return info
}
