package handler

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync/atomic"
	"time"
)

type requestIDKey struct{}

var requestSequence atomic.Uint64

func RequestID(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey{}).(string)
	return id
}

func requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := fmt.Sprintf("req-%d", requestSequence.Add(1))

		w.Header().Set("X-Request-ID", id)

		ctx := context.WithValue(r.Context(), requestIDKey{}, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *Handler) requestLoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusWriter{ResponseWriter: w}
		completed := false

		defer func() {
			status := rec.status
			if status == 0 && completed {
				status = http.StatusOK
			}

			level := slog.LevelInfo
			if status >= 500 || !completed {
				level = slog.LevelError
			}

			h.logger.Log(r.Context(), level, "request",
				"request_id", RequestID(r.Context()), "method", r.Method,
				"path", r.URL.Path, "status", status,
				"duration", time.Since(start).String(),
				"aborted", !completed)
		}()

		next.ServeHTTP(rec, r)
		completed = true
	})
}

func (h *Handler) recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if value := recover(); value != nil {
				if value == http.ErrAbortHandler {
					panic(value)
				}

				h.logger.ErrorContext(r.Context(), "panic recovered", "panic", value, "request_id", RequestID(r.Context()))

				if rec, ok := w.(*statusWriter); ok && rec.status != 0 {
					panic(http.ErrAbortHandler)
				}

				w.Header().Del("Location")
				h.writeJSON(w, r, http.StatusInternalServerError, errorResponse{
					Error: "internal server error",
					Code:  "INTERNAL_ERROR",
				})
			}
		}()
		next.ServeHTTP(w, r)
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	if code >= 100 && code < 200 {
		w.ResponseWriter.WriteHeader(code)
		return
	}
	if w.status != 0 {
		return
	}

	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusWriter) Write(data []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}

	return w.ResponseWriter.Write(data)
}

func (w *statusWriter) Unwrap() http.ResponseWriter { return w.ResponseWriter }
