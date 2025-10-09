package input

import (
	"context"
	"zpwoot/internal/core/application/dto"
)

type ChatwootUseCases interface {
	Create() ChatwootCreateUseCase
	Get() ChatwootGetUseCase
	Update() ChatwootUpdateUseCase
	Delete() ChatwootDeleteUseCase
}

type ChatwootCreateUseCase interface {
	Execute(ctx context.Context, sessionID string, req *dto.CreateChatwootRequest) (*dto.ChatwootResponse, error)
}

type ChatwootGetUseCase interface {
	Execute(ctx context.Context, sessionID string, baseURL string) (*dto.ChatwootResponse, error)
	GetByID(ctx context.Context, id string, baseURL string) (*dto.ChatwootResponse, error)
	List(ctx context.Context, limit, offset int, baseURL string) (*dto.ChatwootListResponse, error)
	ListEnabled(ctx context.Context, limit, offset int, baseURL string) (*dto.ChatwootListResponse, error)
	Exists(ctx context.Context, sessionID string) (bool, error)
}

type ChatwootUpdateUseCase interface {
	Execute(ctx context.Context, sessionID string, req *dto.UpdateChatwootRequest, baseURL string) (*dto.ChatwootResponse, error)
}

type ChatwootDeleteUseCase interface {
	Execute(ctx context.Context, sessionID string) error
}

type ChatwootWebhookHandler interface {
	HandleWebhook(ctx context.Context, sessionID string, req *dto.ChatwootWebhookRequest) (*dto.ChatwootWebhookResponse, error)
}
