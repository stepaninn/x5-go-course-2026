package handler

import (
	"errors"
	"net/http"

	"example.com/order-service/internal/order"
)

// @Summary Check service health
// @Tags health
// @Produce json
// @Success 200 {object} map[string]string "Service status: status = ok"
// @Router /health [get]
func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	h.writeJSON(w, r, http.StatusOK, map[string]string{"status": "ok"})
}

// @Summary Create order
// @Description Create order with status new and calculate total price
// @Tags orders
// @Accept json
// @Produce json
// @Param input body order.CreateOrderInput true "Items in order"
// @Success 201 {object} order.Order "Order created"
// @Failure 400 {object} errorResponse "Invalid JSON (INVALID_JSON) or invalid order (INVALID_ORDER)"
// @Failure 500 {object} errorResponse "Internal server error (INTERNAL_ERROR)"
// @Router /api/v1/orders [post]
func (h *Handler) createOrder(w http.ResponseWriter, r *http.Request) {
	var input order.CreateOrderInput

	if err := decodeJSON(w, r, &input); err != nil {
		h.writeJSON(w, r, err.status, errorResponse{
			Error: err.message,
			Code:  "INVALID_JSON",
		})
		return
	}

	o, err := h.service.Create(r.Context(), input)
	if errors.Is(err, order.ErrInvalidOrder) {
		h.writeJSON(w, r, http.StatusBadRequest, errorResponse{
			Error: "invalid order",
			Code:  "INVALID_ORDER",
		})
		return
	}

	if err != nil {
		h.logger.ErrorContext(r.Context(), "create order failed", "error", err)

		h.writeJSON(w, r, http.StatusInternalServerError, errorResponse{
			Error: "internal server error",
			Code:  "INTERNAL_ERROR",
		})
		return
	}

	h.writeJSON(w, r, http.StatusCreated, o)
}
