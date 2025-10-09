package cache

import (
	"context"
	"fmt"
	"time"

	"zpwoot/internal/core/domain/webhook"
	"zpwoot/internal/core/ports/output"
)

type CachedWebhookRepository struct {
	repository webhook.Repository
	cache      output.CachePort
	keyBuilder output.CacheKeyBuilder
	serializer output.CacheSerializer
	logger     output.Logger
	config     *CacheRepositoryConfig
}

func NewCachedWebhookRepository(
	repository webhook.Repository,
	cache output.CachePort,
	keyBuilder output.CacheKeyBuilder,
	serializer output.CacheSerializer,
	logger output.Logger,
	config *CacheRepositoryConfig,
) webhook.Repository {
	if config == nil {
		config = &CacheRepositoryConfig{
			SessionTTL: 10 * time.Minute,
			ListTTL:    2 * time.Minute,
			Enabled:    true,
		}
	}

	return &CachedWebhookRepository{
		repository: repository,
		cache:      cache,
		keyBuilder: keyBuilder,
		serializer: serializer,
		logger:     logger,
		config:     config,
	}
}

func (r *CachedWebhookRepository) Create(ctx context.Context, wh *webhook.Webhook) error {

	err := r.repository.Create(ctx, wh)
	if err != nil {
		return err
	}

	if !r.config.Enabled {
		return nil
	}

	if err := r.cacheWebhookWithMultipleKeys(ctx, wh); err != nil {
		r.logger.Warn().Err(err).Str("webhook_id", wh.ID).Msg("Failed to cache webhook after create")
	}

	r.invalidateListCache(ctx)

	return nil
}

func (r *CachedWebhookRepository) GetByID(ctx context.Context, id string) (*webhook.Webhook, error) {
	if !r.config.Enabled {
		return r.repository.GetByID(ctx, id)
	}

	key := r.keyBuilder.WebhookKey(id)

	if wh, err := r.getWebhookFromCache(ctx, key); err == nil {
		r.logger.Debug().Str("webhook_id", id).Msg("Webhook found in cache")
		return wh, nil
	}

	wh, err := r.repository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := r.cacheWebhookWithMultipleKeys(ctx, wh); err != nil {
		r.logger.Warn().Err(err).Str("webhook_id", id).Msg("Failed to cache webhook")
	}

	r.logger.Debug().Str("webhook_id", id).Msg("Webhook loaded from database and cached")
	return wh, nil
}

func (r *CachedWebhookRepository) GetBySessionID(ctx context.Context, sessionID string) (*webhook.Webhook, error) {
	if !r.config.Enabled {
		return r.repository.GetBySessionID(ctx, sessionID)
	}

	key := r.keyBuilder.WebhookBySessionKey(sessionID)

	if wh, err := r.getWebhookFromCache(ctx, key); err == nil {
		r.logger.Debug().Str("session_id", sessionID).Msg("Webhook found in cache by session ID")
		return wh, nil
	}

	wh, err := r.repository.GetBySessionID(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	if err := r.cacheWebhookWithMultipleKeys(ctx, wh); err != nil {
		r.logger.Warn().Err(err).Str("session_id", sessionID).Msg("Failed to cache webhook")
	}

	r.logger.Debug().Str("session_id", sessionID).Msg("Webhook loaded from database and cached")
	return wh, nil
}

func (r *CachedWebhookRepository) Update(ctx context.Context, wh *webhook.Webhook) error {

	err := r.repository.Update(ctx, wh)
	if err != nil {
		return err
	}

	if !r.config.Enabled {
		return nil
	}

	r.invalidateWebhookCache(ctx, wh)

	if err := r.cacheWebhookWithMultipleKeys(ctx, wh); err != nil {
		r.logger.Warn().Err(err).Str("webhook_id", wh.ID).Msg("Failed to cache webhook after update")
	}

	r.invalidateListCache(ctx)

	return nil
}

func (r *CachedWebhookRepository) Delete(ctx context.Context, id string) error {

	wh, err := r.GetByID(ctx, id)
	if err == nil && r.config.Enabled {
		r.invalidateWebhookCache(ctx, wh)
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

func (r *CachedWebhookRepository) DeleteBySessionID(ctx context.Context, sessionID string) error {

	wh, err := r.GetBySessionID(ctx, sessionID)
	if err == nil && r.config.Enabled {
		r.invalidateWebhookCache(ctx, wh)
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

func (r *CachedWebhookRepository) List(ctx context.Context, limit, offset int) ([]*webhook.Webhook, error) {
	if !r.config.Enabled {
		return r.repository.List(ctx, limit, offset)
	}

	key := r.keyBuilder.WebhookListKey(limit, offset)

	if webhooks, err := r.getWebhookListFromCache(ctx, key); err == nil {
		r.logger.Debug().Int("limit", limit).Int("offset", offset).Msg("Webhook list found in cache")
		return webhooks, nil
	}

	webhooks, err := r.repository.List(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	if err := r.cacheWebhookList(ctx, key, webhooks); err != nil {
		r.logger.Warn().Err(err).Int("limit", limit).Int("offset", offset).Msg("Failed to cache webhook list")
	}

	r.logger.Debug().Int("limit", limit).Int("offset", offset).Msg("Webhook list loaded from database and cached")
	return webhooks, nil
}

func (r *CachedWebhookRepository) cacheWebhookWithMultipleKeys(ctx context.Context, wh *webhook.Webhook) error {
	keys := []string{
		r.keyBuilder.WebhookKey(wh.ID),
		r.keyBuilder.WebhookBySessionKey(wh.SessionID),
	}

	data, err := r.serializer.Serialize(wh)
	if err != nil {
		return fmt.Errorf("failed to serialize webhook: %w", err)
	}

	items := make(map[string][]byte)
	for _, key := range keys {
		items[key] = data
	}

	return r.cache.MSet(ctx, items, r.config.SessionTTL)
}

func (r *CachedWebhookRepository) getWebhookFromCache(ctx context.Context, key string) (*webhook.Webhook, error) {
	data, err := r.cache.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	var wh webhook.Webhook
	if err := r.serializer.Deserialize(data, &wh); err != nil {
		return nil, fmt.Errorf("failed to deserialize webhook: %w", err)
	}

	return &wh, nil
}

func (r *CachedWebhookRepository) invalidateWebhookCache(ctx context.Context, wh *webhook.Webhook) {
	keys := []string{
		r.keyBuilder.WebhookKey(wh.ID),
		r.keyBuilder.WebhookBySessionKey(wh.SessionID),
	}

	if err := r.cache.MDelete(ctx, keys); err != nil {
		r.logger.Warn().Err(err).Str("webhook_id", wh.ID).Msg("Failed to invalidate webhook cache")
	}
}

func (r *CachedWebhookRepository) cacheWebhookList(ctx context.Context, key string, webhooks []*webhook.Webhook) error {
	data, err := r.serializer.Serialize(webhooks)
	if err != nil {
		return fmt.Errorf("failed to serialize webhook list: %w", err)
	}

	return r.cache.Set(ctx, key, data, r.config.ListTTL)
}

func (r *CachedWebhookRepository) getWebhookListFromCache(ctx context.Context, key string) ([]*webhook.Webhook, error) {
	data, err := r.cache.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	var webhooks []*webhook.Webhook
	if err := r.serializer.Deserialize(data, &webhooks); err != nil {
		return nil, fmt.Errorf("failed to deserialize webhook list: %w", err)
	}

	return webhooks, nil
}

func (r *CachedWebhookRepository) invalidateListCache(ctx context.Context) {
	pattern := r.keyBuilder.BuildKey("webhook", "list", "*")
	if err := r.cache.DeleteByPattern(ctx, pattern); err != nil {
		r.logger.Warn().Err(err).Msg("Failed to invalidate webhook list cache")
	}
}
