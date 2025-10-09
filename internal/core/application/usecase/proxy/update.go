package proxy

import (
	"context"

	"zpwoot/internal/core/application/dto"
	"zpwoot/internal/core/domain/proxy"
	"zpwoot/internal/core/ports/output"
)

type UpdateUseCase struct {
	proxyService *proxy.Service
	logger       output.Logger
}

func NewUpdateUseCase(proxyService *proxy.Service, logger output.Logger) *UpdateUseCase {
	return &UpdateUseCase{
		proxyService: proxyService,
		logger:       logger,
	}
}

func (uc *UpdateUseCase) Execute(ctx context.Context, sessionID string, req *dto.UpdateProxyRequest) (*dto.ProxyResponse, error) {
	// Get existing configuration
	config, err := uc.proxyService.GetBySessionID(ctx, sessionID)
	if err != nil {
		uc.logger.Error().
			Err(err).
			Str("session_id", sessionID).
			Msg("Failed to get existing proxy configuration")
		return nil, err
	}

	// Update fields if provided
	if req.Host != nil && req.Port != nil && req.Protocol != nil {
		config.Update(*req.Host, *req.Port, proxy.Protocol(*req.Protocol))
	} else {
		// Update individual fields
		if req.Host != nil {
			config.Host = *req.Host
		}
		if req.Port != nil {
			config.Port = *req.Port
		}
		if req.Protocol != nil {
			config.Protocol = proxy.Protocol(*req.Protocol)
		}
	}

	if req.Username != nil && req.Password != nil {
		config.SetCredentials(*req.Username, *req.Password)
	}

	if req.Enabled != nil {
		if *req.Enabled {
			config.Enable()
		} else {
			config.Disable()
		}
	}

	// Save updated configuration
	if err := uc.proxyService.Update(ctx, config); err != nil {
		uc.logger.Error().
			Err(err).
			Str("session_id", sessionID).
			Msg("Failed to update proxy configuration")
		return nil, err
	}

	uc.logger.Info().
		Str("session_id", sessionID).
		Str("host", config.Host).
		Int("port", config.Port).
		Str("protocol", string(config.Protocol)).
		Msg("Proxy configuration updated successfully")

	return uc.toResponse(config), nil
}

func (uc *UpdateUseCase) toResponse(config *proxy.ProxyConfig) *dto.ProxyResponse {
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
