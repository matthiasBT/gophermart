package usecases

import (
	"github.com/go-chi/chi/v5"
	"github.com/matthiasBT/gophermart/internal/infra/logging"
	"github.com/matthiasBT/gophermart/internal/server/entities"
)

type BaseController struct {
	logger      logging.ILogger
	stor        entities.Storage
	userRepo    entities.UserRepo
	orderRepo   entities.OrderRepo
	accrualRepo entities.AccrualRepo
	crypto      entities.ICryptoProvider
}

func NewBaseController(
	logger logging.ILogger,
	stor entities.Storage,
	userRepo entities.UserRepo,
	orderRepo entities.OrderRepo,
	accrualRepo entities.AccrualRepo,
	crypto entities.ICryptoProvider,
) *BaseController {
	return &BaseController{
		logger:      logger,
		stor:        stor,
		userRepo:    userRepo,
		orderRepo:   orderRepo,
		accrualRepo: accrualRepo,
		crypto:      crypto,
	}
}

func (c *BaseController) Route() *chi.Mux {
	r := chi.NewRouter()
	r.Post("/user/register", c.register)
	r.Post("/user/login", c.signIn)
	r.Post("/user/orders", c.createOrder)
	r.Get("/user/orders", c.getOrders)
	r.Get("/user/balance", c.getBalance)
	r.Post("/user/balance/withdraw", c.withdraw)
	r.Get("/user/withdrawals", c.getWithdrawals)
	return r
}
