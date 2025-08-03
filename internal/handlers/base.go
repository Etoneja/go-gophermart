package handlers

import (
	"github.com/etoneja/go-gophermart/internal/service"
	"github.com/rs/zerolog"
)

type Handlers struct {
	svc    service.Servicer
	logger zerolog.Logger
}

func NewHandlers(svc service.Servicer, logger zerolog.Logger) *Handlers {
	return &Handlers{
		svc:    svc,
		logger: logger,
	}
}
