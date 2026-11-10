package httpapi

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"example.com/lesson8-practice/internal/repository"
)

func health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func createOrder(store *repository.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Прочитайте customer из JSON и проверьте имя до вызова store.Create.
		// Как завершить обработку сразу после ошибки?
		var input struct {
			Customer string `json:"customer"`
		}

		err := json.NewDecoder(r.Body).Decode(&input)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "bad request body"})
			return
		}

		customer := strings.TrimSpace(input.Customer)
		if customer == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "customer must be non empty"})
			return
		}

		order := store.Create(customer)

		writeJSON(w, http.StatusCreated, order)
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	data, err := json.Marshal(value)
	if err != nil {
		slog.Error("encode response", "error", err)
		status = http.StatusInternalServerError
		data = []byte(`{"error":"internal server error"}`)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err := w.Write(append(data, '\n')); err != nil {
		slog.Error("write response", "error", err)
	}
}
