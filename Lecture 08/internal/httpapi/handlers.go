package httpapi

import (
	"encoding/json"
	"log"
	"net/http"

	"example.com/lesson8-practice/internal/repository"
)

func health(w http.ResponseWriter, r *http.Request) {
	// Как задать формат и статус ответа?
	http.Error(w, "TODO: health", http.StatusNotImplemented)
}

func createOrder(store *repository.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Прочитайте customer из JSON и проверьте имя до вызова store.Create.
		// Как завершить обработку сразу после ошибки?
		http.Error(w, "TODO: create order", http.StatusNotImplemented)
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	data, err := json.Marshal(value)
	if err != nil {
		log.Printf("encode response: %v", err)
		status = http.StatusInternalServerError
		data = []byte(`{"error":"internal server error"}`)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err := w.Write(append(data, '\n')); err != nil {
		log.Printf("write response: %v", err)
	}
}
