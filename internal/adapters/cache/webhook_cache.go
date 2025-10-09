package cache

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"zpwoot/internal/core/domain/webhook"
	"zpwoot/internal/core/ports/output"
)

// CachedWebhook represents a cached webhook configuration with expiration
type CachedWebhook struct {
	Config    *webhook.Webhook
	ExpiresAt time.Time
}

// WebhookCache is an in-memory cache implementation for webhook configurations
// It implements the output.WebhookCache port
type WebhookCache struct {
	cache      map[string]*CachedWebhook
	mu         sync.RWMutex
	ttl        time.Duration
	maxSize    int
	hits       int64
	misses     int64
	totalGetNs int64
	getCount   int64
}

// NewWebhookCache creates a new webhook cache with specified TTL and max size
func NewWebhookCache(ttl time.Duration, maxSize int) *WebhookCache {
	cache := &WebhookCache{
		cache:   make(map[string]*CachedWebhook),
		ttl:     ttl,
		maxSize: maxSize,
	}

	// Start background cleanup goroutine
	go cache.cleanupExpired()

	return cache
}

// Get retrieves a webhook config from cache
func (c *WebhookCache) Get(ctx context.Context, sessionID string) (*webhook.Webhook, error) {
	start := time.Now()
	defer func() {
		atomic.AddInt64(&c.totalGetNs, int64(time.Since(start)))
		atomic.AddInt64(&c.getCount, 1)
	}()

	c.mu.RLock()
	defer c.mu.RUnlock()

	cached, ok := c.cache[sessionID]
	if !ok {
		atomic.AddInt64(&c.misses, 1)
		return nil, nil
	}

	// Check if expired
	if time.Now().After(cached.ExpiresAt) {
		atomic.AddInt64(&c.misses, 1)
		return nil, nil
	}

	atomic.AddInt64(&c.hits, 1)
	return cached.Config, nil
}

// Set stores a webhook config in cache with TTL
func (c *WebhookCache) Set(ctx context.Context, sessionID string, config *webhook.Webhook) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check max size (simple eviction: reject new entries)
	if len(c.cache) >= c.maxSize {
		_, exists := c.cache[sessionID]
		if !exists {
			// Cache is full and this is a new entry
			return nil // Silently ignore (could also evict LRU)
		}
	}

	c.cache[sessionID] = &CachedWebhook{
		Config:    config,
		ExpiresAt: time.Now().Add(c.ttl),
	}

	return nil
}

// Invalidate removes a webhook config from cache
func (c *WebhookCache) Invalidate(ctx context.Context, sessionID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.cache, sessionID)
	return nil
}

// Clear removes all entries from cache
func (c *WebhookCache) Clear(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache = make(map[string]*CachedWebhook)
	return nil
}

// GetMetrics returns cache performance metrics
func (c *WebhookCache) GetMetrics() output.CacheMetrics {
	c.mu.RLock()
	size := len(c.cache)
	c.mu.RUnlock()

	hits := atomic.LoadInt64(&c.hits)
	misses := atomic.LoadInt64(&c.misses)
	totalGetNs := atomic.LoadInt64(&c.totalGetNs)
	getCount := atomic.LoadInt64(&c.getCount)

	total := hits + misses
	hitRate := 0.0
	if total > 0 {
		hitRate = float64(hits) / float64(total) * 100
	}

	avgGetTime := time.Duration(0)
	if getCount > 0 {
		avgGetTime = time.Duration(totalGetNs / getCount)
	}

	return output.CacheMetrics{
		Hits:       hits,
		Misses:     misses,
		Size:       size,
		HitRate:    hitRate,
		AvgGetTime: avgGetTime,
	}
}

// cleanupExpired removes expired entries periodically
func (c *WebhookCache) cleanupExpired() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for sessionID, cached := range c.cache {
			if now.After(cached.ExpiresAt) {
				delete(c.cache, sessionID)
			}
		}
		c.mu.Unlock()
	}
}

