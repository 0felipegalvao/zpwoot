package cache

import (
	"fmt"
	"strings"

	"zpwoot/internal/core/ports/output"
)

type DefaultKeyBuilder struct {
	prefix string
}

func NewDefaultKeyBuilder(prefix string) output.CacheKeyBuilder {
	if prefix == "" {
		prefix = "zpwoot"
	}
	return &DefaultKeyBuilder{
		prefix: prefix,
	}
}

func (kb *DefaultKeyBuilder) BuildKey(parts ...string) string {
	allParts := append([]string{kb.prefix}, parts...)
	return strings.Join(allParts, ":")
}

func (kb *DefaultKeyBuilder) SessionKey(sessionID string) string {
	return kb.BuildKey("session", "id", sessionID)
}

func (kb *DefaultKeyBuilder) SessionByNameKey(name string) string {
	return kb.BuildKey("session", "name", name)
}

func (kb *DefaultKeyBuilder) SessionByJIDKey(jid string) string {
	return kb.BuildKey("session", "jid", jid)
}

func (kb *DefaultKeyBuilder) SessionListKey(limit, offset int) string {
	return kb.BuildKey("session", "list", fmt.Sprintf("%d:%d", limit, offset))
}

func (kb *DefaultKeyBuilder) WebhookKey(webhookID string) string {
	return kb.BuildKey("webhook", "id", webhookID)
}

func (kb *DefaultKeyBuilder) WebhookBySessionKey(sessionID string) string {
	return kb.BuildKey("webhook", "session", sessionID)
}

func (kb *DefaultKeyBuilder) WebhookListKey(limit, offset int) string {
	return kb.BuildKey("webhook", "list", fmt.Sprintf("%d:%d", limit, offset))
}

func (kb *DefaultKeyBuilder) ProxyKey(sessionID string) string {
	return kb.BuildKey("proxy", "session", sessionID)
}

func (kb *DefaultKeyBuilder) ChatwootKey(chatwootID string) string {
	return kb.BuildKey("chatwoot", "id", chatwootID)
}

func (kb *DefaultKeyBuilder) ChatwootBySessionKey(sessionID string) string {
	return kb.BuildKey("chatwoot", "session", sessionID)
}

func (kb *DefaultKeyBuilder) ChatwootListKey(limit, offset int) string {
	return kb.BuildKey("chatwoot", "list", fmt.Sprintf("%d:%d", limit, offset))
}

func (kb *DefaultKeyBuilder) ChatwootListByEnabledKey(enabled bool, limit, offset int) string {
	enabledStr := "false"
	if enabled {
		enabledStr = "true"
	}
	return kb.BuildKey("chatwoot", "list", "enabled", enabledStr, fmt.Sprintf("%d:%d", limit, offset))
}

func (kb *DefaultKeyBuilder) SessionPattern() string {
	return kb.BuildKey("session", "*")
}

func (kb *DefaultKeyBuilder) WebhookPattern() string {
	return kb.BuildKey("webhook", "*")
}

func (kb *DefaultKeyBuilder) ProxyPattern() string {
	return kb.BuildKey("proxy", "*")
}

func (kb *DefaultKeyBuilder) ChatwootPattern() string {
	return kb.BuildKey("chatwoot", "*")
}

func (kb *DefaultKeyBuilder) SessionStatusKey(sessionID string) string {
	return kb.BuildKey("session", "status", sessionID)
}

func (kb *DefaultKeyBuilder) SessionQRKey(sessionID string) string {
	return kb.BuildKey("session", "qr", sessionID)
}

func (kb *DefaultKeyBuilder) SessionConfigKey(sessionID string) string {
	return kb.BuildKey("session", "config", sessionID)
}

func (kb *DefaultKeyBuilder) WebhookEventsKey(sessionID string) string {
	return kb.BuildKey("webhook", "events", sessionID)
}

func (kb *DefaultKeyBuilder) UserSessionsKey(userID string) string {
	return kb.BuildKey("user", "sessions", userID)
}

func (kb *DefaultKeyBuilder) ActiveSessionsKey() string {
	return kb.BuildKey("sessions", "active")
}

func (kb *DefaultKeyBuilder) ConnectedSessionsKey() string {
	return kb.BuildKey("sessions", "connected")
}

func (kb *DefaultKeyBuilder) SessionStatsKey(sessionID string) string {
	return kb.BuildKey("session", "stats", sessionID)
}

func (kb *DefaultKeyBuilder) GlobalStatsKey() string {
	return kb.BuildKey("stats", "global")
}

func (kb *DefaultKeyBuilder) HealthCheckKey() string {
	return kb.BuildKey("health", "check")
}

func (kb *DefaultKeyBuilder) LockKey(resource string) string {
	return kb.BuildKey("lock", resource)
}

func (kb *DefaultKeyBuilder) TempKey(purpose string, identifier string) string {
	return kb.BuildKey("temp", purpose, identifier)
}

func (kb *DefaultKeyBuilder) MetricsKey(metric string) string {
	return kb.BuildKey("metrics", metric)
}

func (kb *DefaultKeyBuilder) ConfigKey(configName string) string {
	return kb.BuildKey("config", configName)
}

func (kb *DefaultKeyBuilder) RateLimitKey(identifier string) string {
	return kb.BuildKey("ratelimit", identifier)
}

func (kb *DefaultKeyBuilder) GetPrefix() string {
	return kb.prefix
}

func (kb *DefaultKeyBuilder) ValidateKey(key string) bool {
	return strings.HasPrefix(key, kb.prefix+":")
}

func (kb *DefaultKeyBuilder) ExtractParts(key string) []string {
	if !kb.ValidateKey(key) {
		return nil
	}

	withoutPrefix := strings.TrimPrefix(key, kb.prefix+":")
	return strings.Split(withoutPrefix, ":")
}

func (kb *DefaultKeyBuilder) GetKeyType(key string) string {
	parts := kb.ExtractParts(key)
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}
