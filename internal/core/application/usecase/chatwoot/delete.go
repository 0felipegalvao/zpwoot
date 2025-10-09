package chatwoot

import (
	"context"
	"fmt"

	"zpwoot/internal/core/domain/chatwoot"
	"zpwoot/internal/core/ports/output"
)

type DeleteUseCase struct {
	chatwootService *chatwoot.Service
	logger          output.Logger
}

func NewDeleteUseCase(chatwootService *chatwoot.Service, logger output.Logger, baseURL string) *DeleteUseCase {
	return &DeleteUseCase{
		chatwootService: chatwootService,
		logger:          logger,
	}
}

func (uc *DeleteUseCase) Execute(ctx context.Context, sessionID string) error {

	exists, err := uc.chatwootService.ConfigurationExists(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to check if chatwoot configuration exists: %w", err)
	}

	if !exists {
		uc.logger.Warn().
			Str("session_id", sessionID).
			Msg("Chatwoot configuration not found for deletion")
		return nil
	}

	config, err := uc.chatwootService.GetConfiguration(ctx, sessionID)
	if err != nil {
		uc.logger.Warn().
			Str("session_id", sessionID).
			Err(err).
			Msg("Failed to get chatwoot configuration for logging, proceeding with deletion")
	}

	err = uc.chatwootService.DeleteConfiguration(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete chatwoot configuration: %w", err)
	}

	if config != nil {
		uc.logger.Info().
			Str("session_id", sessionID).
			Str("chatwoot_id", config.ID).
			Msg("Chatwoot configuration deleted successfully")
	} else {
		uc.logger.Info().
			Str("session_id", sessionID).
			Msg("Chatwoot configuration deleted successfully")
	}

	return nil
}
