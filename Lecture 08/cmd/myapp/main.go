package main

import (
	"log"
	"net/http"
	"time"

	"example.com/lesson8-practice/internal/httpapi"
	"example.com/lesson8-practice/internal/repository"
)

func main() {
	server := &http.Server{
		Addr: "127.0.0.1:8080", Handler: httpapi.NewRouter(repository.NewStore()),
		ReadHeaderTimeout: 3 * time.Second,
		ReadTimeout:       10 * time.Second, WriteTimeout: 10 * time.Second,
		IdleTimeout: 60 * time.Second,
	}
	log.Fatal(server.ListenAndServe())
}
