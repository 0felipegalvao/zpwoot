package chatwoot

import (
	"context"
	"fmt"

	"zpwoot/internal/core/application/dto"
	"zpwoot/internal/core/application/validator"
	"zpwoot/internal/core/domain/chatwoot"
	"zpwoot/internal/core/domain/common"
	"zpwoot/internal/core/ports/output"
)

type CreateUseCase struct {
	chatwootService *chatwoot.Service
	logger          output.Logger
	baseURL         string
}

func NewCreateUseCase(chatwootService *chatwoot.Service, logger output.Logger, baseURL string) *CreateUseCase {
	return &CreateUseCase{
		chatwootService: chatwootService,
		logger:          logger,
		baseURL:         baseURL,
	}
}

func (uc *CreateUseCase) Execute(ctx context.Context, sessionID string, req *dto.CreateChatwootRequest) (*dto.ChatwootResponse, error) {

	if _, err := validator.ValidateStruct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	config, err := uc.chatwootService.CreateConfiguration(ctx, sessionID, req.URL, req.Token, req.AccountID)
	if err != nil {
		if err == common.ErrAlreadyExists {
			return nil, fmt.Errorf("chatwoot configuration already exists for session")
		}
		return nil, fmt.Errorf("failed to create chatwoot configuration: %w", err)
	}

	if !enabled {
		config, err = uc.chatwootService.DisableConfiguration(ctx, sessionID)
		if err != nil {
			return nil, fmt.Errorf("failed to disable chatwoot configuration: %w", err)
		}
	}

	if hasAdvancedSettings(req) {
		settings := buildAdvancedSettings(req)
		config, err = uc.chatwootService.UpdateAdvancedSettings(ctx, sessionID, settings)
		if err != nil {
			return nil, fmt.Errorf("failed to update advanced settings: %w", err)
		}
	}

	if req.InboxID != nil {
		config, err = uc.chatwootService.UpdateConfiguration(ctx, sessionID, config.URL, config.Token, config.AccountID, req.InboxID)
		if err != nil {
			return nil, fmt.Errorf("failed to update inbox ID: %w", err)
		}
	}

	uc.logger.Info().
		Str("session_id", sessionID).
		Str("chatwoot_id", config.ID).
		Bool("enabled", config.Enabled).
		Msg("Chatwoot configuration created successfully")

	return uc.toResponse(config, uc.baseURL), nil
}

func hasAdvancedSettings(req *dto.CreateChatwootRequest) bool {
	return req.InboxName != nil ||
		req.AutoCreate != nil ||
		req.SignMsg != nil ||
		req.SignDelimiter != nil ||
		req.ReopenConv != nil ||
		req.ConvPending != nil ||
		req.ImportContacts != nil ||
		req.ImportMessages != nil ||
		req.ImportDays != nil ||
		req.MergeBrazil != nil ||
		req.Organization != nil ||
		req.Logo != nil ||
		req.Number != nil ||
		len(req.IgnoreJids) > 0
}

func buildAdvancedSettings(req *dto.CreateChatwootRequest) *chatwoot.AdvancedSettings {
	settings := &chatwoot.AdvancedSettings{}

	if req.InboxName != nil {
		settings.InboxName = req.InboxName
	}
	if req.AutoCreate != nil {
		settings.AutoCreate = req.AutoCreate
	}
	if req.SignMsg != nil {
		settings.SignMsg = req.SignMsg
	}
	if req.SignDelimiter != nil {
		settings.SignDelimiter = req.SignDelimiter
	}
	if req.ReopenConv != nil {
		settings.ReopenConv = req.ReopenConv
	}
	if req.ConvPending != nil {
		settings.ConvPending = req.ConvPending
	}
	if req.ImportContacts != nil {
		settings.ImportContacts = req.ImportContacts
	}
	if req.ImportMessages != nil {
		settings.ImportMessages = req.ImportMessages
	}
	if req.ImportDays != nil {
		settings.ImportDays = req.ImportDays
	}
	if req.MergeBrazil != nil {
		settings.MergeBrazil = req.MergeBrazil
	}
	if req.Organization != nil {
		settings.Organization = req.Organization
	}
	if req.Logo != nil {
		settings.Logo = req.Logo
	}
	if req.Number != nil {
		settings.Number = req.Number
	}
	if len(req.IgnoreJids) > 0 {
		settings.IgnoreJids = &req.IgnoreJids
	}

	return settings
}

func (uc *CreateUseCase) toResponse(config *chatwoot.Chatwoot, baseURL string) *dto.ChatwootResponse {
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

func maskToken(token string) string {
	if len(token) <= 8 {
		return "****"
	}
	return token[:4] + "****" + token[len(token)-4:]
}
