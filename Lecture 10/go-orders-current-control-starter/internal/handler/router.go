package handler

import (
	"context"
	"log/slog"
	"net/http"

	"example.com/order-service/internal/order"
	"github.com/go-chi/chi/v5"
)

type OrderService interface {
	Create(ctx context.Context, input order.CreateOrderInput) (order.Order, error)
	GetByID(ctx context.Context, id int64) (order.Order, error)
	List(ctx context.Context) ([]order.Order, error)
	UpdateStatus(ctx context.Context, id int64, status order.Status) (order.Order, error)
	History(ctx context.Context, id int64) ([]order.StatusChange, error)
}

type Handler struct {
	service OrderService
	logger  *slog.Logger
}

func NewRouter(service OrderService, logger *slog.Logger) http.Handler {
	h := &Handler{service: service, logger: logger}
	router := chi.NewRouter()
	router.Use(requestIDMiddleware, h.requestLoggingMiddleware, h.recoveryMiddleware)

	router.Get("/health", h.health)
	router.Route("/api/v1/orders", func(r chi.Router) {
		r.Post("/", h.createOrder)
		r.Get("/", h.listOrders)
		r.Get("/{id}", h.getOrderByID)
		r.Patch("/{id}/status", h.updateStatus)
		r.Get("/{id}/history", h.getOrderHistory)
	})

	return router
}

type errorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}
