package handler

import (
	"errors"
	"net/http"
	"strconv"

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
		h.logger.ErrorContext(r.Context(), "create order failed", "error", err, "request_id", RequestID(r.Context()))

		h.writeJSON(w, r, http.StatusInternalServerError, errorResponse{
			Error: "internal server error",
			Code:  "INTERNAL_ERROR",
		})
		return
	}

	h.writeJSON(w, r, http.StatusCreated, o)
}

// @Summary Get order by ID
// @Description Return order by its ID
// @Tags orders
// @Produce json
// @Param id path int true "Order ID" minimum(1)
// @Success 200 {object} order.Order "Order found"
// @Failure 400 {object} errorResponse "Invalid order ID (INVALID_ID)"
// @Failure 404 {object} errorResponse "Order not found (ORDER_NOT_FOUND)"
// @Failure 500 {object} errorResponse "Internal server error (INTERNAL_ERROR)"
// @Router /api/v1/orders/{id} [get]
func (h *Handler) getOrderByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		h.writeJSON(w, r, http.StatusBadRequest, errorResponse{
			Error: "id must be a positive integer",
			Code:  "INVALID_ID",
		})
		return
	}

	o, err := h.service.GetByID(r.Context(), int64(id))
	if errors.Is(err, order.ErrOrderNotFound) {
		h.writeJSON(w, r, http.StatusNotFound, errorResponse{
			Error: "order not found",
			Code:  "ORDER_NOT_FOUND",
		})
		return
	}

	if err != nil {
		h.logger.ErrorContext(r.Context(), "get order failed", "error", err, "request_id", RequestID(r.Context()))

		h.writeJSON(w, r, http.StatusInternalServerError, errorResponse{
			Error: "internal server error",
			Code:  "INTERNAL_ERROR",
		})
		return
	}

	h.writeJSON(w, r, http.StatusOK, o)
}

// @Summary List orders
// @Description Return all orders sorted by creation time
// @Tags orders
// @Produce json
// @Success 200 {array} order.Order "Orders"
// @Failure 500 {object} errorResponse "Internal server error (INTERNAL_ERROR)"
// @Router /api/v1/orders [get]
func (h *Handler) listOrders(w http.ResponseWriter, r *http.Request) {
	orders, err := h.service.List(r.Context())
	if err != nil {
		h.logger.ErrorContext(r.Context(), "list orders failed", "error", err, "request_id", RequestID(r.Context()))

		h.writeJSON(w, r, http.StatusInternalServerError, errorResponse{
			Error: "internal server error",
			Code:  "INTERNAL_ERROR",
		})
		return
	}

	h.writeJSON(w, r, http.StatusOK, orders)
}

// @Summary Update order status
// @Description Allowed transitions: new to processing or cancelled; processing to completed or cancelled. Completed and cancelled are final statuses.
// @Tags orders
// @Accept json
// @Produce json
// @Param id path int true "Order ID" minimum(1)
// @Param input body order.UpdateStatusInput true "New order status"
// @Success 200 {object} order.Order "Order updated"
// @Failure 400 {object} errorResponse "Invalid ID (INVALID_ID), JSON (INVALID_JSON) or status (INVALID_STATUS)"
// @Failure 404 {object} errorResponse "Order not found (ORDER_NOT_FOUND)"
// @Failure 409 {object} errorResponse "Status transition is not allowed (INVALID_STATUS_TRANSITION)"
// @Failure 500 {object} errorResponse "Internal server error (INTERNAL_ERROR)"
// @Router /api/v1/orders/{id}/status [patch]
func (h *Handler) updateStatus(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		h.writeJSON(w, r, http.StatusBadRequest, errorResponse{
			Error: "id must be a positive integer",
			Code:  "INVALID_ID",
		})
		return
	}

	var input order.UpdateStatusInput

	if err := decodeJSON(w, r, &input); err != nil {
		h.writeJSON(w, r, err.status, errorResponse{
			Error: err.message,
			Code:  "INVALID_JSON",
		})
		return
	}

	o, err := h.service.UpdateStatus(r.Context(), int64(id), input.Status)
	if errors.Is(err, order.ErrOrderNotFound) {
		h.writeJSON(w, r, http.StatusNotFound, errorResponse{
			Error: "order not found",
			Code:  "ORDER_NOT_FOUND",
		})
		return
	}

	if errors.Is(err, order.ErrInvalidStatus) {
		h.writeJSON(w, r, http.StatusBadRequest, errorResponse{
			Error: "invalid status",
			Code:  "INVALID_STATUS",
		})
		return
	}

	if errors.Is(err, order.ErrInvalidStatusTransition) {
		h.writeJSON(w, r, http.StatusConflict, errorResponse{
			Error: "status transition is not allowed",
			Code:  "INVALID_STATUS_TRANSITION",
		})
		return
	}

	if err != nil {
		h.logger.ErrorContext(r.Context(), "update status failed", "error", err, "request_id", RequestID(r.Context()))

		h.writeJSON(w, r, http.StatusInternalServerError, errorResponse{
			Error: "internal server error",
			Code:  "INTERNAL_ERROR",
		})
		return
	}

	h.writeJSON(w, r, http.StatusOK, o)
}

// @Summary Get order history
// @Description Return status changes in chronological order
// @Tags orders
// @Produce json
// @Param id path int true "Order ID" minimum(1)
// @Success 200 {array} order.StatusChange "Order history"
// @Failure 400 {object} errorResponse "Invalid order ID (INVALID_ID)"
// @Failure 404 {object} errorResponse "Order not found (ORDER_NOT_FOUND)"
// @Failure 500 {object} errorResponse "Internal server error (INTERNAL_ERROR)"
// @Router /api/v1/orders/{id}/history [get]
func (h *Handler) getOrderHistory(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		h.writeJSON(w, r, http.StatusBadRequest, errorResponse{
			Error: "id must be a positive integer",
			Code:  "INVALID_ID",
		})
		return
	}

	history, err := h.service.History(r.Context(), int64(id))
	if errors.Is(err, order.ErrOrderNotFound) {
		h.writeJSON(w, r, http.StatusNotFound, errorResponse{
			Error: "order not found",
			Code:  "ORDER_NOT_FOUND",
		})
		return
	}

	if err != nil {
		h.logger.ErrorContext(r.Context(), "get order history failed", "error", err, "request_id", RequestID(r.Context()))

		h.writeJSON(w, r, http.StatusInternalServerError, errorResponse{
			Error: "internal server error",
			Code:  "INTERNAL_ERROR",
		})
		return
	}

	h.writeJSON(w, r, http.StatusOK, history)
}
