package cache

import (
	"context"
	"time"

	"zpwoot/internal/core/ports/output"
)

type NoOpCache struct{}

func NewNoOpCache() output.CachePort {
	return &NoOpCache{}
}

func (c *NoOpCache) Get(ctx context.Context, key string) ([]byte, error) {
	return nil, output.ErrCacheKeyNotFound
}

func (c *NoOpCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return nil
}

func (c *NoOpCache) Delete(ctx context.Context, key string) error {
	return nil
}

func (c *NoOpCache) Exists(ctx context.Context, key string) (bool, error) {
	return false, nil
}

func (c *NoOpCache) MGet(ctx context.Context, keys []string) (map[string][]byte, error) {
	return make(map[string][]byte), nil
}

func (c *NoOpCache) MSet(ctx context.Context, items map[string][]byte, ttl time.Duration) error {
	return nil
}

func (c *NoOpCache) MDelete(ctx context.Context, keys []string) error {
	return nil
}

func (c *NoOpCache) Keys(ctx context.Context, pattern string) ([]string, error) {
	return []string{}, nil
}

func (c *NoOpCache) DeleteByPattern(ctx context.Context, pattern string) error {
	return nil
}

func (c *NoOpCache) Clear(ctx context.Context) error {
	return nil
}

func (c *NoOpCache) FlushDB(ctx context.Context) error {
	return nil
}

func (c *NoOpCache) Ping(ctx context.Context) error {
	return nil
}

func (c *NoOpCache) Info(ctx context.Context) (map[string]string, error) {
	return map[string]string{
		"type":    "noop",
		"status":  "ok",
		"enabled": "false",
	}, nil
}

func (c *NoOpCache) Close() error {
	return nil
}
