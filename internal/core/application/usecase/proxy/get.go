package proxy

import (
	"context"

	"zpwoot/internal/core/application/dto"
	"zpwoot/internal/core/domain/proxy"
	"zpwoot/internal/core/ports/output"
)

type GetUseCase struct {
	proxyService *proxy.Service
	logger       output.Logger
}

func NewGetUseCase(proxyService *proxy.Service, logger output.Logger) *GetUseCase {
	return &GetUseCase{
		proxyService: proxyService,
		logger:       logger,
	}
}

func (uc *GetUseCase) Execute(ctx context.Context, sessionID string) (*dto.ProxyResponse, error) {
	config, err := uc.proxyService.GetBySessionID(ctx, sessionID)
	if err != nil {
		uc.logger.Error().
			Err(err).
			Str("session_id", sessionID).
			Msg("Failed to get proxy configuration")
		return nil, err
	}

	uc.logger.Debug().
		Str("session_id", sessionID).
		Msg("Proxy configuration retrieved successfully")

	return uc.toResponse(config), nil
}

func (uc *GetUseCase) toResponse(config *proxy.ProxyConfig) *dto.ProxyResponse {
	return &dto.ProxyResponse{
		ID:        config.ID,
		SessionID: config.SessionID,
		Host:      config.Host,
		Port:      config.Port,
		Protocol:  string(config.Protocol),
		Username:  config.Username,
		Password:  config.Password,
		Enabled:   config.Enabled,
		URL:       config.GetURL(),
		CreatedAt: config.CreatedAt,
		UpdatedAt: config.UpdatedAt,
	}
}
