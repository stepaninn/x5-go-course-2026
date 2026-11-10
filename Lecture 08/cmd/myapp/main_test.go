package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"example.com/lesson8-practice/internal/httpapi"
	"example.com/lesson8-practice/internal/repository"
)

// До реализации health этот тест должен быть красным.
func TestHealth(t *testing.T) {
	h := httpapi.NewRouter(repository.NewStore())
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d, want 200", rec.Code)
	}
	if rec.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("Content-Type=%q", rec.Header().Get("Content-Type"))
	}
	var body struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Status != "ok" {
		t.Fatalf("status in body=%q", body.Status)
	}
}

// Добавьте проверки успешного создания и ошибочного входа.
