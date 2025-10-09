package chatwoot

import "context"

type Repository interface {
	Create(ctx context.Context, chatwoot *Chatwoot) error

	GetByID(ctx context.Context, id string) (*Chatwoot, error)

	GetBySessionID(ctx context.Context, sessionID string) (*Chatwoot, error)

	Update(ctx context.Context, chatwoot *Chatwoot) error

	Delete(ctx context.Context, id string) error

	DeleteBySessionID(ctx context.Context, sessionID string) error

	List(ctx context.Context, limit, offset int) ([]*Chatwoot, error)

	ListByEnabled(ctx context.Context, enabled bool, limit, offset int) ([]*Chatwoot, error)

	Exists(ctx context.Context, sessionID string) (bool, error)
}
