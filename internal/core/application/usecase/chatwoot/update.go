package chatwoot

import (
	"context"
	"fmt"

	"zpwoot/internal/core/application/dto"
	"zpwoot/internal/core/application/validator"
	"zpwoot/internal/core/domain/chatwoot"
	"zpwoot/internal/core/ports/output"
)

type UpdateUseCase struct {
	chatwootService *chatwoot.Service
	logger          output.Logger
	baseURL         string
}

func NewUpdateUseCase(chatwootService *chatwoot.Service, logger output.Logger, baseURL string) *UpdateUseCase {
	return &UpdateUseCase{
		chatwootService: chatwootService,
		logger:          logger,
		baseURL:         baseURL,
	}
}

func (uc *UpdateUseCase) Execute(ctx context.Context, sessionID string, req *dto.UpdateChatwootRequest, baseURL string) (*dto.ChatwootResponse, error) {

	if _, err := validator.ValidateStruct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	config, err := uc.chatwootService.GetConfiguration(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get chatwoot configuration: %w", err)
	}

	if req.URL != nil || req.Token != nil || req.AccountID != nil || req.InboxID != nil {
		url := config.URL
		token := config.Token
		accountID := config.AccountID
		inboxID := config.InboxID

		if req.URL != nil {
			url = *req.URL
		}
		if req.Token != nil {
			token = *req.Token
		}
		if req.AccountID != nil {
			accountID = *req.AccountID
		}
		if req.InboxID != nil {
			inboxID = req.InboxID
		}

		config, err = uc.chatwootService.UpdateConfiguration(ctx, sessionID, url, token, accountID, inboxID)
		if err != nil {
			return nil, fmt.Errorf("failed to update chatwoot configuration: %w", err)
		}
	}

	if req.Enabled != nil {
		if *req.Enabled {
			config, err = uc.chatwootService.EnableConfiguration(ctx, sessionID)
		} else {
			config, err = uc.chatwootService.DisableConfiguration(ctx, sessionID)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to update enabled status: %w", err)
		}
	}

	if hasAdvancedSettingsUpdate(req) {
		settings := buildAdvancedSettingsUpdate(req)
		config, err = uc.chatwootService.UpdateAdvancedSettings(ctx, sessionID, settings)
		if err != nil {
			return nil, fmt.Errorf("failed to update advanced settings: %w", err)
		}
	}

	uc.logger.Info().
		Str("session_id", sessionID).
		Str("chatwoot_id", config.ID).
		Bool("enabled", config.Enabled).
		Msg("Chatwoot configuration updated successfully")

	return uc.toResponse(config, uc.baseURL), nil
}

func hasAdvancedSettingsUpdate(req *dto.UpdateChatwootRequest) bool {
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
		req.IgnoreJids != nil
}

func buildAdvancedSettingsUpdate(req *dto.UpdateChatwootRequest) *chatwoot.AdvancedSettings {
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
	if req.IgnoreJids != nil {
		settings.IgnoreJids = req.IgnoreJids
	}

	return settings
}

func (uc *UpdateUseCase) toResponse(config *chatwoot.Chatwoot, baseURL string) *dto.ChatwootResponse {
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
