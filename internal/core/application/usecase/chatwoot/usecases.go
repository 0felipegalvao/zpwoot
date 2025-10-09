package chatwoot

import (
	"zpwoot/internal/core/domain/chatwoot"
	"zpwoot/internal/core/ports/input"
	"zpwoot/internal/core/ports/output"
)

type UseCases struct {
	create *CreateUseCase
	get    *GetUseCase
	update *UpdateUseCase
	delete *DeleteUseCase
}

func NewUseCases(chatwootService *chatwoot.Service, logger output.Logger, baseURL string) input.ChatwootUseCases {
	return &UseCases{
		create: NewCreateUseCase(chatwootService, logger, baseURL),
		get:    NewGetUseCase(chatwootService, logger, baseURL),
		update: NewUpdateUseCase(chatwootService, logger, baseURL),
		delete: NewDeleteUseCase(chatwootService, logger, baseURL),
	}
}

func (uc *UseCases) Create() input.ChatwootCreateUseCase {
	return uc.create
}

func (uc *UseCases) Get() input.ChatwootGetUseCase {
	return uc.get
}

func (uc *UseCases) Update() input.ChatwootUpdateUseCase {
	return uc.update
}

func (uc *UseCases) Delete() input.ChatwootDeleteUseCase {
	return uc.delete
}
