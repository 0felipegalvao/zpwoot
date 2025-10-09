package input

import (
	"context"

	"zpwoot/internal/core/application/dto"
)

type ProxyCreateUseCase interface {
	Execute(ctx context.Context, sessionID string, req *dto.CreateProxyRequest) (*dto.ProxyResponse, error)
}

type ProxyGetUseCase interface {
	Execute(ctx context.Context, sessionID string) (*dto.ProxyResponse, error)
}

type ProxyUpdateUseCase interface {
	Execute(ctx context.Context, sessionID string, req *dto.UpdateProxyRequest) (*dto.ProxyResponse, error)
}

type ProxyUseCases interface {
	Create() ProxyCreateUseCase
	Get() ProxyGetUseCase
	Update() ProxyUpdateUseCase
}
