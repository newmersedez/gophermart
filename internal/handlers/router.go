package handlers

import (
	"log/slog"
	"net/http"

	"gophermart/internal/middleware"
	"gophermart/internal/storage"

	"github.com/go-chi/chi/v5"
)

type Router struct {
	authHandler    *AuthHandler
	orderHandler   *OrderHandler
	balanceHandler *BalanceHandler
}

func NewRouter(storage *storage.Storage, logger *slog.Logger) *Router {
	return &Router{
		authHandler:    NewAuthHandler(storage, logger),
		orderHandler:   NewOrderHandler(storage, logger),
		balanceHandler: NewBalanceHandler(storage, logger),
	}
}

func (rt *Router) Routes(logger *slog.Logger) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestLoggerMiddleware(logger))

	r.Route("/api/user", func(r chi.Router) {
		r.Post("/register", rt.authHandler.Register)
		r.Post("/login", rt.authHandler.Login)

		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthMiddleware)

			r.Post("/orders", rt.orderHandler.UploadOrder)
			r.Get("/orders", rt.orderHandler.GetOrders)

			r.Get("/balance", rt.balanceHandler.GetBalance)
			r.Post("/balance/withdraw", rt.balanceHandler.Withdraw)

			r.Get("/withdrawals", rt.balanceHandler.GetWithdrawals)
		})
	})

	return r
}
