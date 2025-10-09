package proxy

import "context"

type Repository interface {
	Create(ctx context.Context, config *ProxyConfig) error
	GetBySessionID(ctx context.Context, sessionID string) (*ProxyConfig, error)
	Update(ctx context.Context, config *ProxyConfig) error
	Delete(ctx context.Context, sessionID string) error
}
