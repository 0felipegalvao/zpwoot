package cache

import (
	"context"
	"fmt"
	"strings"
	"time"

	"zpwoot/internal/core/ports/output"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
	config *output.CacheConfig
	logger output.Logger
}

func NewRedisCache(config *output.CacheConfig, logger output.Logger) (output.CachePort, error) {
	if config == nil {
		return nil, fmt.Errorf("cache config is required")
	}

	var opts *redis.Options
	if config.URL != "" {
		parsedOpts, err := redis.ParseURL(config.URL)
		if err != nil {
			return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
		}
		opts = parsedOpts
	} else {
		opts = &redis.Options{
			Addr:     "localhost:6379",
			Password: config.Password,
			DB:       config.DB,
		}
	}

	if config.DialTimeout > 0 {
		opts.DialTimeout = config.DialTimeout
	}
	if config.ReadTimeout > 0 {
		opts.ReadTimeout = config.ReadTimeout
	}
	if config.WriteTimeout > 0 {
		opts.WriteTimeout = config.WriteTimeout
	}

	if config.PoolSize > 0 {
		opts.PoolSize = config.PoolSize
	}
	if config.MinIdleConns > 0 {
		opts.MinIdleConns = config.MinIdleConns
	}

	client := redis.NewClient(opts)

	cache := &RedisCache{
		client: client,
		config: config,
		logger: logger,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := cache.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.Info().Str("addr", opts.Addr).Msg("Redis cache connected successfully")

	return cache, nil
}

func (r *RedisCache) Get(ctx context.Context, key string) ([]byte, error) {
	result, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, output.ErrCacheKeyNotFound
		}
		r.logger.Error().Err(err).Str("key", key).Msg("Redis GET failed")
		return nil, fmt.Errorf("redis get failed: %w", err)
	}

	return []byte(result), nil
}

func (r *RedisCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if ttl <= 0 {
		ttl = r.config.DefaultTTL
	}

	err := r.client.Set(ctx, key, value, ttl).Err()
	if err != nil {
		r.logger.Error().Err(err).Str("key", key).Dur("ttl", ttl).Msg("Redis SET failed")
		return fmt.Errorf("redis set failed: %w", err)
	}

	return nil
}

func (r *RedisCache) Delete(ctx context.Context, key string) error {
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		r.logger.Error().Err(err).Str("key", key).Msg("Redis DELETE failed")
		return fmt.Errorf("redis delete failed: %w", err)
	}

	return nil
}

func (r *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	result, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		r.logger.Error().Err(err).Str("key", key).Msg("Redis EXISTS failed")
		return false, fmt.Errorf("redis exists failed: %w", err)
	}

	return result > 0, nil
}

func (r *RedisCache) MGet(ctx context.Context, keys []string) (map[string][]byte, error) {
	if len(keys) == 0 {
		return make(map[string][]byte), nil
	}

	results, err := r.client.MGet(ctx, keys...).Result()
	if err != nil {
		r.logger.Error().Err(err).Strs("keys", keys).Msg("Redis MGET failed")
		return nil, fmt.Errorf("redis mget failed: %w", err)
	}

	data := make(map[string][]byte)
	for i, result := range results {
		if result != nil {
			if str, ok := result.(string); ok {
				data[keys[i]] = []byte(str)
			}
		}
	}

	return data, nil
}

func (r *RedisCache) MSet(ctx context.Context, items map[string][]byte, ttl time.Duration) error {
	if len(items) == 0 {
		return nil
	}

	if ttl <= 0 {
		ttl = r.config.DefaultTTL
	}

	pipe := r.client.Pipeline()

	for key, value := range items {
		pipe.Set(ctx, key, value, ttl)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		r.logger.Error().Err(err).Int("count", len(items)).Msg("Redis MSET failed")
		return fmt.Errorf("redis mset failed: %w", err)
	}

	return nil
}

func (r *RedisCache) MDelete(ctx context.Context, keys []string) error {
	if len(keys) == 0 {
		return nil
	}

	err := r.client.Del(ctx, keys...).Err()
	if err != nil {
		r.logger.Error().Err(err).Strs("keys", keys).Msg("Redis MDEL failed")
		return fmt.Errorf("redis mdel failed: %w", err)
	}

	return nil
}

func (r *RedisCache) Keys(ctx context.Context, pattern string) ([]string, error) {
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		r.logger.Error().Err(err).Str("pattern", pattern).Msg("Redis KEYS failed")
		return nil, fmt.Errorf("redis keys failed: %w", err)
	}

	return keys, nil
}

func (r *RedisCache) DeleteByPattern(ctx context.Context, pattern string) error {
	keys, err := r.Keys(ctx, pattern)
	if err != nil {
		return err
	}

	if len(keys) == 0 {
		return nil
	}

	return r.MDelete(ctx, keys)
}

func (r *RedisCache) Clear(ctx context.Context) error {
	pattern := r.config.KeyPrefix + "*"
	return r.DeleteByPattern(ctx, pattern)
}

func (r *RedisCache) FlushDB(ctx context.Context) error {
	err := r.client.FlushDB(ctx).Err()
	if err != nil {
		r.logger.Error().Err(err).Msg("Redis FLUSHDB failed")
		return fmt.Errorf("redis flushdb failed: %w", err)
	}

	return nil
}

func (r *RedisCache) Ping(ctx context.Context) error {
	err := r.client.Ping(ctx).Err()
	if err != nil {
		return fmt.Errorf("redis ping failed: %w", err)
	}

	return nil
}

func (r *RedisCache) Info(ctx context.Context) (map[string]string, error) {
	result, err := r.client.Info(ctx).Result()
	if err != nil {
		r.logger.Error().Err(err).Msg("Redis INFO failed")
		return nil, fmt.Errorf("redis info failed: %w", err)
	}

	info := make(map[string]string)
	lines := strings.Split(result, "\r\n")
	for _, line := range lines {
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				info[parts[0]] = parts[1]
			}
		}
	}

	return info, nil
}

func (r *RedisCache) Close() error {
	if r.client != nil {
		err := r.client.Close()
		if err != nil {
			r.logger.Error().Err(err).Msg("Failed to close Redis connection")
			return err
		}
		r.logger.Info().Msg("Redis connection closed")
	}
	return nil
}
