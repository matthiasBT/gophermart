package usecases

import (
	"github.com/go-chi/chi/v5"
	"github.com/matthiasBT/gophermart/internal/infra/logging"
	"github.com/matthiasBT/gophermart/internal/server/entities"
)

type BaseController struct {
	Logger logging.ILogger
	Stor   entities.Storage
}

func NewBaseController(logger logging.ILogger, stor entities.Storage) *BaseController {
	return &BaseController{
		Logger: logger,
		Stor:   stor,
	}
}

func (c *BaseController) Route() *chi.Mux {
	r := chi.NewRouter()
	r.Post("/api/user/register", c.register)
	return r
}
