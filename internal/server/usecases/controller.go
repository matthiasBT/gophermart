package usecases

import (
	"github.com/go-chi/chi/v5"
	"github.com/matthiasBT/gophermart/internal/infra/logging"
	"github.com/matthiasBT/gophermart/internal/server/entities"
)

type BaseController struct {
	logger  logging.ILogger
	stor    entities.Storage
	crypto  entities.ICryptoProvider
	accrual entities.IAccrualClient
}

func NewBaseController(logger logging.ILogger, stor entities.Storage, crypto entities.ICryptoProvider, accrual entities.IAccrualClient) *BaseController {
	return &BaseController{
		logger:  logger,
		stor:    stor,
		crypto:  crypto,
		accrual: accrual,
	}
}

func (c *BaseController) Route() *chi.Mux {
	r := chi.NewRouter()
	r.Post("/api/user/register", c.register) // todo: mount with prefix
	r.Post("/api/user/login", c.signIn)
	r.Post("/api/user/orders", c.createOrder)
	r.Get("/api/user/orders", c.getOrders)
	r.Get("/api/user/balance", c.getBalance)
	r.Post("/api/user/balance/withdraw", c.withdraw)
	return r
}
