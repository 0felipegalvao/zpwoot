package proxy

import (
	"context"

	"zpwoot/internal/core/application/dto"
	"zpwoot/internal/core/domain/proxy"
	"zpwoot/internal/core/ports/output"
)

type CreateUseCase struct {
	proxyService *proxy.Service
	logger       output.Logger
}

func NewCreateUseCase(proxyService *proxy.Service, logger output.Logger) *CreateUseCase {
	return &CreateUseCase{
		proxyService: proxyService,
		logger:       logger,
	}
}

func (uc *CreateUseCase) Execute(ctx context.Context, sessionID string, req *dto.CreateProxyRequest) (*dto.ProxyResponse, error) {
	// Set defaults
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	// Create proxy config
	config := proxy.NewProxyConfig(sessionID, req.Host, req.Port, proxy.Protocol(req.Protocol))
	config.Enabled = enabled

	if req.Username != nil && req.Password != nil {
		config.SetCredentials(*req.Username, *req.Password)
	}

	// Save to repository
	if err := uc.proxyService.Create(ctx, config); err != nil {
		uc.logger.Error().
			Err(err).
			Str("session_id", sessionID).
			Msg("Failed to create proxy configuration")
		return nil, err
	}

	uc.logger.Info().
		Str("session_id", sessionID).
		Str("host", config.Host).
		Int("port", config.Port).
		Str("protocol", string(config.Protocol)).
		Msg("Proxy configuration created successfully")

	return uc.toResponse(config), nil
}

func (uc *CreateUseCase) toResponse(config *proxy.ProxyConfig) *dto.ProxyResponse {
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
