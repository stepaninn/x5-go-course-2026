package httpapi

import (
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"

	"example.com/lesson9practice/internal/model"
	"example.com/lesson9practice/internal/repository"
	"example.com/lesson9practice/internal/service"
)

type Orders interface {
	Create(string) (model.Order, error)
	ByID(int) (model.Order, error)
	ChangeStatusByID(int, string) (model.Order, error)
	List(model.ListOptions) ([]model.Order, error)
}

type API struct {
	orders Orders
	logger *slog.Logger
}

func New(orders Orders, logger *slog.Logger) *API { return &API{orders: orders, logger: logger} }

// @summary Check service status
// @tags health
// @produce json
// @success 200 {object} StatusResponse
// @router /status [get]
func (a *API) status(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, StatusResponse{Status: "ok"})
}

// @summary Create order with given name
// @tags orders
// @accept json
// @produce json
// @param request body CreateOrderRequest true "Data for order creating"
// @success 201 {object} OrderResponse
// @header 201 {string} Location "Location to the created order"
// @failure 400 {object} ErrorResponse
// @failure 413 {object} ErrorResponse
// @failure 415 {object} ErrorResponse
// @failure 500 {object} ErrorResponse
// @router /orders [post]
func (a *API) createOrder(w http.ResponseWriter, r *http.Request) {
	// Указатель позволяет отличить null от объекта. Теги DTO сами не валидируют запрос.
	var in *CreateOrderRequest
	if err := decodeJSON(w, r, &in); err != nil {
		writeError(w, err.status, err.message, err.code)
		return
	}
	if in == nil {
		writeError(w, 400, "body must be a JSON object", "INVALID_JSON")
		return
	}
	order, err := a.orders.Create(in.Customer)
	if errors.Is(err, service.ErrEmptyCustomer) {
		writeError(w, 400, "customer is required", "VALIDATION_ERROR")
		return
	}
	if err != nil {
		a.internalError(w, r, err)
		return
	}
	w.Header().Set("Location", "/api/v1/orders/"+strconv.Itoa(order.ID))
	writeJSON(w, 201, orderResponse(order))
}

// @summary Get order by ID
// @tags orders
// @produce json
// @param id path int true "Order ID" minimum(1)
// @success 200 {object} OrderResponse
// @failure 400 {object} ErrorResponse
// @failure 404 {object} ErrorResponse
// @failure 500 {object} ErrorResponse
// @router /orders/{id} [get]
func (a *API) getOrder(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		writeError(w, 400, "id must be a positive integer", "INVALID_ID")
		return
	}
	order, err := a.orders.ByID(id)
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, 404, "order not found", "ORDER_NOT_FOUND")
		return
	}
	if err != nil {
		a.internalError(w, r, err)
		return
	}
	writeJSON(w, 200, orderResponse(order))
}

// @summary Change order status by ID
// @tags orders
// @accept json
// @produce json
// @param id path int true "Order ID" minimum(1)
// @param request body ChangeStatusRequest true "Data for status changing"
// @success 200 {object} OrderResponse
// @failure 400 {object} ErrorResponse
// @failure 404 {object} ErrorResponse
// @failure 409 {object} ErrorResponse
// @failure 413 {object} ErrorResponse
// @failure 415 {object} ErrorResponse
// @failure 500 {object} ErrorResponse
// @router /orders/{id}/status [patch]
func (a *API) changeStatus(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		writeError(w, 400, "id must be a positive integer", "INVALID_ID")
		return
	}

	var in *ChangeStatusRequest
	if err := decodeJSON(w, r, &in); err != nil {
		writeError(w, err.status, err.message, err.code)
		return
	}
	if in == nil {
		writeError(w, 400, "body must be a JSON object", "INVALID_JSON")
		return
	}
	if !model.ValidateStatus(in.Status) {
		writeError(w, 400, "invalid status", "INVALID_JSON")
		return
	}

	order, err := a.orders.ChangeStatusByID(id, in.Status)
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, 404, "order not found", "ORDER_NOT_FOUND")
		return
	}
	if errors.Is(err, repository.ErrAlreadyCancelled) {
		writeError(w, 409, "order already cancelled", "ORDER_ALREADY_CANCELLED")
		return
	}
	if errors.Is(err, repository.ErrInvalidStatusTransition) {
		writeError(w, 400, "unsupported status change", "UNSUPPORTED_STATUS_CHANGE")
		return
	}
	if err != nil {
		a.internalError(w, r, err)
		return
	}

	writeJSON(w, 200, orderResponse(order))
}

// @summary List orders with query parameters
// @tags orders
// @produce json
// @param limit query int false "Limit" minimum(1) maximum(100) default(20)
// @param offset query int false "Offset" minimum(0) default(0)
// @param status query string false "Status" enums(new, cancelled)
// @success 200 {array} OrderResponse
// @failure 400 {object} ErrorResponse
// @failure 500 {object} ErrorResponse
// @router /orders [get]
func (a *API) listOrders(w http.ResponseWriter, r *http.Request) {
	values, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		writeError(w, 400, "invalid query", "INVALID_QUERY")
		return
	}

	options, ok := parseQueryOptions(values)
	if !ok {
		writeError(w, 400, "invalid query", "INVALID_QUERY")
		return
	}

	orders, err := a.orders.List(options)
	if err != nil {
		a.internalError(w, r, err)
		return
	}

	body := make([]OrderResponse, 0, len(orders))
	for _, o := range orders {
		body = append(body, orderResponse(o))
	}

	writeJSON(w, 200, body)
}

func parseQueryOptions(v url.Values) (model.ListOptions, bool) {
	options := model.ListOptions{Limit: 20}
	for key, items := range v {
		switch key {
		case "limit", "offset":
			if len(items) != 1 || items[0] == "" {
				return options, false
			}

			for _, c := range items[0] {
				if c < '0' || c > '9' {
					return options, false
				}
			}

			number, err := strconv.Atoi(items[0])
			if err != nil || number < 0 {
				return options, false
			}

			if key == "limit" {
				if number < 1 || number > 100 {
					return options, false
				}
				options.Limit = number
			} else {
				options.Offset = number
			}
		case "status":
			if len(items) != 1 {
				return options, false
			}

			if status := items[0]; !model.ValidateStatus(status) {
				return options, false
			}
			options.Status = items[0]
		default:
			return options, false
		}
	}
	return options, true
}

func (a *API) internalError(w http.ResponseWriter, r *http.Request, err error) {
	a.logger.ErrorContext(r.Context(), "order operation failed", "error", err)
	writeError(w, 500, "internal server error", "INTERNAL_ERROR")
}
