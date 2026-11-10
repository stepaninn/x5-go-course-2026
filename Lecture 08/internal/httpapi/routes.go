package httpapi

import (
	"net/http"

	"example.com/lesson8-practice/internal/repository"

	"github.com/go-chi/chi/v5"
)

func NewRouter(store *repository.Store) http.Handler {
	router := chi.NewRouter()
	router.Use(logging)
	router.Get("/health", health)
	router.Head("/health", health)
	router.Get("/orders/{id}", getOrder(store))
	router.Head("/orders/{id}", getOrder(store))
	router.Post("/orders", createOrder(store))
	return router
}
