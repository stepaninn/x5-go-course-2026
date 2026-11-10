package httpapi_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"example.com/lesson8-practice/internal/httpapi"
	"example.com/lesson8-practice/internal/repository"
	"github.com/go-chi/chi/v5"
)

func TestRecoveryBeforeResponse(t *testing.T) {
	var logs bytes.Buffer
	previousLogger := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&logs, nil)))
	defer slog.SetDefault(previousLogger)

	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("internal panic details")
	})

	store := repository.NewStore()
	handler := httpapi.NewRouter(store)

	router, ok := handler.(chi.Router)
	if !ok {
		t.Fatal("NewRouter did not return a chi router")
	}

	router.Get("/test-panic", panicHandler)
	request := httptest.NewRequest(http.MethodGet, "/test-panic", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("status=%d, want 500; body=%s", response.Code, response.Body.String())
	}
	if got := response.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type=%q, want application/json", got)
	}

	var body struct {
		Error string `json:"error"`
	}

	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if body.Error != "internal server error" {
		t.Errorf("error=%q, want internal server error", body.Error)
	}
	if strings.Contains(response.Body.String(), "internal panic details") {
		t.Error("response exposes panic details")
	}

	accessLogCount := 0
	for _, line := range strings.Split(strings.TrimSpace(logs.String()), "\n") {
		var entry struct {
			Message string `json:"msg"`
			Status  int    `json:"status"`
			Method  string `json:"method"`
			Path    string `json:"path"`
		}

		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			t.Fatalf("decode log entry: %v; line=%s", err, line)
		}
		if entry.Message != "request" {
			continue
		}

		accessLogCount++

		if entry.Status != http.StatusInternalServerError {
			t.Errorf("access log status=%d, want 500", entry.Status)
		}
		if entry.Method != http.MethodGet || entry.Path != "/test-panic" {
			t.Errorf("unexpected access log request: method=%q, path=%q", entry.Method, entry.Path)
		}
	}

	if accessLogCount != 1 {
		t.Errorf("access log count=%d, want 1; logs=%s", accessLogCount, logs.String())
	}
}
