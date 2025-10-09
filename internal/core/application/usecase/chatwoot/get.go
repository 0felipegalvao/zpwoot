package chatwoot

import (
	"context"
	"fmt"

	"zpwoot/internal/core/application/dto"
	"zpwoot/internal/core/domain/chatwoot"
	"zpwoot/internal/core/domain/common"
	"zpwoot/internal/core/ports/output"
)

type GetUseCase struct {
	chatwootService *chatwoot.Service
	logger          output.Logger
	baseURL         string
}

func NewGetUseCase(chatwootService *chatwoot.Service, logger output.Logger, baseURL string) *GetUseCase {
	return &GetUseCase{
		chatwootService: chatwootService,
		logger:          logger,
		baseURL:         baseURL,
	}
}

func (uc *GetUseCase) Execute(ctx context.Context, sessionID string, baseURL string) (*dto.ChatwootResponse, error) {
	config, err := uc.chatwootService.GetConfiguration(ctx, sessionID)
	if err != nil {
		if err == common.ErrNotFound {

			return &dto.ChatwootResponse{
				SessionID:  sessionID,
				Enabled:    false,
				URL:        "",
				AccountID:  "",
				Token:      "",
				WebhookURL: "",
				IgnoreJids: []string{},
			}, nil
		}
		return nil, fmt.Errorf("failed to get chatwoot configuration: %w", err)
	}

	uc.logger.Debug().
		Str("session_id", sessionID).
		Str("chatwoot_id", config.ID).
		Bool("enabled", config.Enabled).
		Msg("Chatwoot configuration retrieved successfully")

	return uc.toResponse(config, uc.baseURL), nil
}

func (uc *GetUseCase) GetByID(ctx context.Context, id string, baseURL string) (*dto.ChatwootResponse, error) {
	config, err := uc.chatwootService.GetConfigurationByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get chatwoot configuration by ID: %w", err)
	}

	uc.logger.Debug().
		Str("chatwoot_id", id).
		Str("session_id", config.SessionID).
		Bool("enabled", config.Enabled).
		Msg("Chatwoot configuration retrieved by ID successfully")

	return uc.toResponse(config, baseURL), nil
}

func (uc *GetUseCase) List(ctx context.Context, limit, offset int, baseURL string) (*dto.ChatwootListResponse, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	configurations, err := uc.chatwootService.ListConfigurations(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list chatwoot configurations: %w", err)
	}

	responses := make([]dto.ChatwootResponse, 0, len(configurations))
	for _, config := range configurations {
		responses = append(responses, *uc.toResponse(config, baseURL))
	}

	uc.logger.Debug().
		Int("count", len(configurations)).
		Int("limit", limit).
		Int("offset", offset).
		Msg("Chatwoot configurations listed successfully")

	return &dto.ChatwootListResponse{
		Configurations: responses,
		Total:          len(responses),
		Limit:          limit,
		Offset:         offset,
	}, nil
}

func (uc *GetUseCase) ListEnabled(ctx context.Context, limit, offset int, baseURL string) (*dto.ChatwootListResponse, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	configurations, err := uc.chatwootService.ListEnabledConfigurations(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list enabled chatwoot configurations: %w", err)
	}

	responses := make([]dto.ChatwootResponse, 0, len(configurations))
	for _, config := range configurations {
		responses = append(responses, *uc.toResponse(config, baseURL))
	}

	uc.logger.Debug().
		Int("count", len(configurations)).
		Int("limit", limit).
		Int("offset", offset).
		Msg("Enabled chatwoot configurations listed successfully")

	return &dto.ChatwootListResponse{
		Configurations: responses,
		Total:          len(responses),
		Limit:          limit,
		Offset:         offset,
	}, nil
}

func (uc *GetUseCase) Exists(ctx context.Context, sessionID string) (bool, error) {
	exists, err := uc.chatwootService.ConfigurationExists(ctx, sessionID)
	if err != nil {
		return false, fmt.Errorf("failed to check if chatwoot configuration exists: %w", err)
	}

	return exists, nil
}

func (uc *GetUseCase) toResponse(config *chatwoot.Chatwoot, baseURL string) *dto.ChatwootResponse {
	response := &dto.ChatwootResponse{
		ID:             config.ID,
		SessionID:      config.SessionID,
		URL:            config.URL,
		Token:          maskToken(config.Token),
		AccountID:      config.AccountID,
		InboxID:        config.InboxID,
		Enabled:        config.Enabled,
		InboxName:      config.InboxName,
		AutoCreate:     config.AutoCreate,
		SignMsg:        config.SignMsg,
		SignDelimiter:  config.SignDelimiter,
		ReopenConv:     config.ReopenConv,
		ConvPending:    config.ConvPending,
		ImportContacts: config.ImportContacts,
		ImportMessages: config.ImportMessages,
		ImportDays:     config.ImportDays,
		MergeBrazil:    config.MergeBrazil,
		Organization:   config.Organization,
		Logo:           config.Logo,
		Number:         config.Number,
		IgnoreJids:     config.IgnoreJids,
		CreatedAt:      config.CreatedAt,
		UpdatedAt:      config.UpdatedAt,
	}

	if baseURL != "" {
		response.WebhookURL = config.GetWebhookURL(baseURL)
	}

	return response
}
