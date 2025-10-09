package output

import (
	"context"
	"time"

	"zpwoot/internal/core/domain/webhook"
)

// WebhookCache defines the interface for webhook configuration caching
// This is an output port that will be implemented by infrastructure adapters
type WebhookCache interface {
	// Get retrieves a webhook config from cache
	// Returns nil if not found or expired
	Get(ctx context.Context, sessionID string) (*webhook.Webhook, error)

	// Set stores a webhook config in cache with TTL
	Set(ctx context.Context, sessionID string, config *webhook.Webhook) error

	// Invalidate removes a webhook config from cache
	Invalidate(ctx context.Context, sessionID string) error

	// Clear removes all entries from cache
	Clear(ctx context.Context) error

	// GetMetrics returns cache performance metrics
	GetMetrics() CacheMetrics
}

// CacheMetrics provides cache performance statistics
type CacheMetrics struct {
	Hits      int64         // Number of cache hits
	Misses    int64         // Number of cache misses
	Size      int           // Current number of entries
	HitRate   float64       // Hit rate percentage (0-100)
	AvgGetTime time.Duration // Average get operation time
}

