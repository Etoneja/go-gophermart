package middlewares

import (
	"github.com/etoneja/go-gophermart/internal/service"
	"github.com/rs/zerolog"
)

type Middlewares struct {
	svc    service.Servicer
	logger zerolog.Logger
}

func NewMiddlewares(svc service.Servicer, logger zerolog.Logger) *Middlewares {
	return &Middlewares{
		svc:    svc,
		logger: logger,
	}
}
