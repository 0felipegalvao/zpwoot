package output

import (
	"context"
	"errors"
	"time"
)

var (
	ErrCacheKeyNotFound = errors.New("cache key not found")
	ErrCacheKeyExpired  = errors.New("cache key expired")
	ErrCacheConnection  = errors.New("cache connection error")
	ErrCacheOperation   = errors.New("cache operation error")
)

type CachePort interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)

	MGet(ctx context.Context, keys []string) (map[string][]byte, error)
	MSet(ctx context.Context, items map[string][]byte, ttl time.Duration) error
	MDelete(ctx context.Context, keys []string) error

	Keys(ctx context.Context, pattern string) ([]string, error)
	DeleteByPattern(ctx context.Context, pattern string) error

	Clear(ctx context.Context) error
	FlushDB(ctx context.Context) error

	Ping(ctx context.Context) error
	Info(ctx context.Context) (map[string]string, error)

	Close() error
}

type CacheKeyBuilder interface {
	SessionKey(sessionID string) string
	SessionByNameKey(name string) string
	SessionByJIDKey(jid string) string
	SessionListKey(limit, offset int) string

	WebhookKey(webhookID string) string
	WebhookBySessionKey(sessionID string) string
	WebhookListKey(limit, offset int) string

	ProxyKey(sessionID string) string

	ChatwootKey(chatwootID string) string
	ChatwootBySessionKey(sessionID string) string
	ChatwootListKey(limit, offset int) string
	ChatwootListByEnabledKey(enabled bool, limit, offset int) string

	BuildKey(parts ...string) string

	SessionPattern() string
	WebhookPattern() string
	ProxyPattern() string
	ChatwootPattern() string
}

type CacheConfig struct {
	URL      string
	Password string
	DB       int

	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration

	PoolSize     int
	MinIdleConns int
	MaxIdleConns int

	DefaultTTL time.Duration
	SessionTTL time.Duration
	WebhookTTL time.Duration
	ProxyTTL   time.Duration
	ListTTL    time.Duration

	KeyPrefix string

	EnableCompression bool
	EnableMetrics     bool
}

type CacheMetrics interface {
	IncrementHits(operation string)
	IncrementMisses(operation string)
	IncrementErrors(operation string)
	RecordLatency(operation string, duration time.Duration)
	GetStats() map[string]interface{}
}

type CacheSerializer interface {
	Serialize(data interface{}) ([]byte, error)
	Deserialize(data []byte, target interface{}) error
}
