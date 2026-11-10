package httpapi

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"example.com/lesson8-practice/internal/repository"
)

func health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, r, http.StatusOK, map[string]string{"status": "ok"})
}

func createOrder(store *repository.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			Customer string `json:"customer"`
		}

		err := json.NewDecoder(r.Body).Decode(&input)
		if err != nil {
			writeJSON(w, r, http.StatusBadRequest, map[string]string{"error": "bad request body"})
			return
		}

		customer := strings.TrimSpace(input.Customer)
		if customer == "" {
			writeJSON(w, r, http.StatusBadRequest, map[string]string{"error": "customer must be non empty"})
			return
		}

		order := store.Create(customer)

		w.Header().Set("Location", "/orders/"+strconv.Itoa(order.ID))
		writeJSON(w, r, http.StatusCreated, order)
	}
}

func getOrder(store *repository.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil || id <= 0 {
			writeJSON(w, r, http.StatusBadRequest, map[string]string{"error": "invalid order id"})
			return
		}

		order, ok := store.ByID(id)
		if !ok {
			writeJSON(w, r, http.StatusNotFound, map[string]string{"error": "order not found"})
			return
		}

		writeJSON(w, r, http.StatusOK, order)
	}
}

func writeJSON(w http.ResponseWriter, r *http.Request, status int, value any) {
	data, err := json.Marshal(value)
	if err != nil {
		slog.Error("encode response",
			"error", err,
			"request-id", RequestID(r.Context()),
		)
		status = http.StatusInternalServerError
		data = []byte(`{"error":"internal server error"}`)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err := w.Write(append(data, '\n')); err != nil {
		slog.Error("write response",
			"error", err,
			"request-id", RequestID(r.Context()),
		)
	}
}
