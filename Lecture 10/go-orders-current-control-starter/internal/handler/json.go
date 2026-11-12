package handler

import (
	"encoding/json"
	"io"
	"net/http"
)

type inputError struct {
	status  int
	message string
}

func decodeJSON(_ http.ResponseWriter, r *http.Request, dst any) *inputError {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		return &inputError{status: http.StatusBadRequest, message: "invalid JSON body"}
	}

	if err := decoder.Decode(new(any)); err != io.EOF {
		return &inputError{http.StatusBadRequest, "body must contain one JSON value"}
	}

	return nil
}

func (h *Handler) writeJSON(w http.ResponseWriter, r *http.Request, status int, value any) {
	data, err := json.Marshal(value)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "encode response failed", "error", err)
		status, data = http.StatusInternalServerError, []byte(`{"error":"internal server error","code":"INTERNAL_ERROR"}`)
		w.Header().Del("Location")
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if _, err := w.Write(append(data, '\n')); err != nil {
		h.logger.ErrorContext(r.Context(), "write response failed", "error", err)
	}
}
