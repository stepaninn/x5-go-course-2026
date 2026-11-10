package httpapi

import (
	"net/http"

	"example.com/lesson8-practice/internal/repository"
)

func NewRouter(store *repository.Store) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", health)
	mux.HandleFunc("POST /orders", createOrder(store))
	return mux
}
