package cache

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"zpwoot/internal/core/ports/output"
)

type MemoryCacheConfig struct {
	DefaultTTL        time.Duration
	CleanupInterval   time.Duration
	MaxSize           int
	EnableCompression bool
	EnableMetrics     bool
}

type MemoryCache struct {
	data    map[string]cacheItem
	mutex   sync.RWMutex
	config  *MemoryCacheConfig
	logger  output.Logger
	metrics *memoryCacheMetrics
	stopCh  chan struct{}
}

type cacheItem struct {
	value     []byte
	expiresAt time.Time
	createdAt time.Time
}

type memoryCacheMetrics struct {
	hits   int64
	misses int64
	sets   int64
	dels   int64
	mutex  sync.RWMutex
}

func NewMemoryCache(logger output.Logger, config *MemoryCacheConfig) output.CachePort {
	if config == nil {
		config = &MemoryCacheConfig{
			DefaultTTL:        5 * time.Minute,
			CleanupInterval:   1 * time.Minute,
			MaxSize:           10000,
			EnableCompression: false,
			EnableMetrics:     true,
		}
	}

	cache := &MemoryCache{
		data:    make(map[string]cacheItem),
		config:  config,
		logger:  logger,
		metrics: &memoryCacheMetrics{},
		stopCh:  make(chan struct{}),
	}

	go cache.cleanup()

	logger.Info().
		Dur("default_ttl", config.DefaultTTL).
		Dur("cleanup_interval", config.CleanupInterval).
		Int("max_size", config.MaxSize).
		Msg("Memory cache initialized")

	return cache
}

func (c *MemoryCache) Get(ctx context.Context, key string) ([]byte, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	item, exists := c.data[key]
	if !exists {
		c.recordMiss()
		return nil, output.ErrCacheKeyNotFound
	}

	if time.Now().After(item.expiresAt) {
		c.recordMiss()

		return nil, output.ErrCacheKeyExpired
	}

	c.recordHit()
	return item.value, nil
}

func (c *MemoryCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if ttl <= 0 {
		ttl = c.config.DefaultTTL
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	if len(c.data) >= c.config.MaxSize {

		c.removeExpiredItemsUnsafe()

		if len(c.data) >= c.config.MaxSize {
			c.removeOldestItemUnsafe()
		}
	}

	c.data[key] = cacheItem{
		value:     value,
		expiresAt: time.Now().Add(ttl),
		createdAt: time.Now(),
	}

	c.recordSet()
	return nil
}

func (c *MemoryCache) Delete(ctx context.Context, key string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	delete(c.data, key)
	c.recordDel()
	return nil
}

func (c *MemoryCache) Exists(ctx context.Context, key string) (bool, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	item, exists := c.data[key]
	if !exists {
		return false, nil
	}

	if time.Now().After(item.expiresAt) {
		return false, nil
	}

	return true, nil
}

func (c *MemoryCache) MGet(ctx context.Context, keys []string) (map[string][]byte, error) {
	if len(keys) == 0 {
		return make(map[string][]byte), nil
	}

	c.mutex.RLock()
	defer c.mutex.RUnlock()

	result := make(map[string][]byte)
	now := time.Now()

	for _, key := range keys {
		if item, exists := c.data[key]; exists && now.Before(item.expiresAt) {
			result[key] = item.value
			c.recordHit()
		} else {
			c.recordMiss()
		}
	}

	return result, nil
}

func (c *MemoryCache) MSet(ctx context.Context, items map[string][]byte, ttl time.Duration) error {
	if len(items) == 0 {
		return nil
	}

	if ttl <= 0 {
		ttl = c.config.DefaultTTL
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	now := time.Now()
	expiresAt := now.Add(ttl)

	for key, value := range items {

		if len(c.data) >= c.config.MaxSize {
			c.removeExpiredItemsUnsafe()
			if len(c.data) >= c.config.MaxSize {
				c.removeOldestItemUnsafe()
			}
		}

		c.data[key] = cacheItem{
			value:     value,
			expiresAt: expiresAt,
			createdAt: now,
		}
		c.recordSet()
	}

	return nil
}

func (c *MemoryCache) MDelete(ctx context.Context, keys []string) error {
	if len(keys) == 0 {
		return nil
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	for _, key := range keys {
		delete(c.data, key)
		c.recordDel()
	}

	return nil
}

func (c *MemoryCache) Keys(ctx context.Context, pattern string) ([]string, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	var keys []string
	now := time.Now()

	for key, item := range c.data {

		if now.After(item.expiresAt) {
			continue
		}

		if c.matchPattern(key, pattern) {
			keys = append(keys, key)
		}
	}

	return keys, nil
}

func (c *MemoryCache) DeleteByPattern(ctx context.Context, pattern string) error {
	keys, err := c.Keys(ctx, pattern)
	if err != nil {
		return err
	}

	return c.MDelete(ctx, keys)
}

func (c *MemoryCache) Clear(ctx context.Context) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.data = make(map[string]cacheItem)
	c.logger.Info().Msg("Memory cache cleared")
	return nil
}

func (c *MemoryCache) FlushDB(ctx context.Context) error {
	return c.Clear(ctx)
}

func (c *MemoryCache) Ping(ctx context.Context) error {
	return nil
}

func (c *MemoryCache) Info(ctx context.Context) (map[string]string, error) {
	c.mutex.RLock()
	totalKeys := len(c.data)
	c.mutex.RUnlock()

	stats := c.getStats()

	return map[string]string{
		"type":        "memory",
		"status":      "ok",
		"total_keys":  fmt.Sprintf("%d", totalKeys),
		"max_size":    fmt.Sprintf("%d", c.config.MaxSize),
		"default_ttl": c.config.DefaultTTL.String(),
		"hits":        fmt.Sprintf("%d", stats["hits"]),
		"misses":      fmt.Sprintf("%d", stats["misses"]),
		"sets":        fmt.Sprintf("%d", stats["sets"]),
		"deletes":     fmt.Sprintf("%d", stats["deletes"]),
		"hit_ratio":   fmt.Sprintf("%.2f", stats["hit_ratio"]),
	}, nil
}

func (c *MemoryCache) Close() error {
	close(c.stopCh)
	c.logger.Info().Msg("Memory cache closed")
	return nil
}

func (c *MemoryCache) cleanup() {
	ticker := time.NewTicker(c.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.removeExpiredItems()
		case <-c.stopCh:
			return
		}
	}
}

func (c *MemoryCache) removeExpiredItems() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.removeExpiredItemsUnsafe()
}

func (c *MemoryCache) removeExpiredItemsUnsafe() {
	now := time.Now()
	for key, item := range c.data {
		if now.After(item.expiresAt) {
			delete(c.data, key)
		}
	}
}

func (c *MemoryCache) removeOldestItemUnsafe() {
	var oldestKey string
	var oldestTime time.Time

	for key, item := range c.data {
		if oldestKey == "" || item.createdAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = item.createdAt
		}
	}

	if oldestKey != "" {
		delete(c.data, oldestKey)
	}
}

func (c *MemoryCache) matchPattern(key, pattern string) bool {

	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(key, prefix)
	}
	return key == pattern
}

func (c *MemoryCache) recordHit() {
	if !c.config.EnableMetrics {
		return
	}
	c.metrics.mutex.Lock()
	c.metrics.hits++
	c.metrics.mutex.Unlock()
}

func (c *MemoryCache) recordMiss() {
	if !c.config.EnableMetrics {
		return
	}
	c.metrics.mutex.Lock()
	c.metrics.misses++
	c.metrics.mutex.Unlock()
}

func (c *MemoryCache) recordSet() {
	if !c.config.EnableMetrics {
		return
	}
	c.metrics.mutex.Lock()
	c.metrics.sets++
	c.metrics.mutex.Unlock()
}

func (c *MemoryCache) recordDel() {
	if !c.config.EnableMetrics {
		return
	}
	c.metrics.mutex.Lock()
	c.metrics.dels++
	c.metrics.mutex.Unlock()
}

func (c *MemoryCache) getStats() map[string]interface{} {
	if !c.config.EnableMetrics {
		return map[string]interface{}{
			"metrics_enabled": false,
		}
	}

	c.metrics.mutex.RLock()
	defer c.metrics.mutex.RUnlock()

	total := c.metrics.hits + c.metrics.misses
	hitRatio := 0.0
	if total > 0 {
		hitRatio = float64(c.metrics.hits) / float64(total)
	}

	return map[string]interface{}{
		"metrics_enabled": true,
		"hits":            c.metrics.hits,
		"misses":          c.metrics.misses,
		"sets":            c.metrics.sets,
		"deletes":         c.metrics.dels,
		"hit_ratio":       hitRatio,
	}
}
