package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

func newRouter(store *Store) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", health)
	mux.HandleFunc("POST /orders", createOrder(store))
	return mux
}

func health(w http.ResponseWriter, r *http.Request) {
	// Как задать формат и статус ответа?
	http.Error(w, "TODO: health", http.StatusNotImplemented)
}

func createOrder(store *Store) http.HandlerFunc {
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

func main() {
	server := &http.Server{
		Addr: "127.0.0.1:8080", Handler: newRouter(NewStore()),
		ReadHeaderTimeout: 3 * time.Second,
		ReadTimeout:       10 * time.Second, WriteTimeout: 10 * time.Second,
		IdleTimeout: 60 * time.Second,
	}
	log.Fatal(server.ListenAndServe())
}
