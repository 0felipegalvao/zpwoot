package proxy

import (
	"zpwoot/internal/core/domain/proxy"
	"zpwoot/internal/core/ports/input"
	"zpwoot/internal/core/ports/output"
)

type UseCases struct {
	create *CreateUseCase
	get    *GetUseCase
	update *UpdateUseCase
}

func NewUseCases(proxyService *proxy.Service, logger output.Logger) input.ProxyUseCases {
	return &UseCases{
		create: NewCreateUseCase(proxyService, logger),
		get:    NewGetUseCase(proxyService, logger),
		update: NewUpdateUseCase(proxyService, logger),
	}
}

func (uc *UseCases) Create() input.ProxyCreateUseCase {
	return uc.create
}

func (uc *UseCases) Get() input.ProxyGetUseCase {
	return uc.get
}

func (uc *UseCases) Update() input.ProxyUpdateUseCase {
	return uc.update
}
