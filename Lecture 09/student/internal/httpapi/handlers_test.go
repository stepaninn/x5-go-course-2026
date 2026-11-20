package httpapi

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net/http"
	"net/http/httptest"
	"slices"
	"strconv"
	"strings"
	"testing"

	"example.com/lesson9practice/internal/model"
	"example.com/lesson9practice/internal/repository"
	"example.com/lesson9practice/internal/service"
)

func newHandler() http.Handler {
	return New(service.New(repository.New()), slog.New(slog.NewTextHandler(io.Discard, nil))).Handler()
}

func request(handler http.Handler, method, path, body, contentType string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	if contentType != "" {
		r.Header.Set("Content-Type", contentType)
	}
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	return w
}

func TestListOrders(t *testing.T) {
	const path = "/api/v1/orders"
	handler := newHandler()
	orders := make([]OrderResponse, 21)
	for i := range orders {
		customer := fmt.Sprintf("Client %d", i+1)
		body := fmt.Sprintf(`{"customer":"Client %d"}`, i+1)

		w := request(handler, http.MethodPost, path, body, "application/json")
		if w.Code != 201 {
			t.Fatalf("create order %d: status=%d body=%s", i+1, w.Code, w.Body)
		}

		orders[i] = OrderResponse{ID: i + 1, Customer: customer, Status: "new"}
	}

	tests := []struct {
		name  string
		query string
		empty bool
		want  []OrderResponse
	}{
		{name: "empty repository", empty: true, want: []OrderResponse{}},
		{name: "default page", want: orders[:20]},
		{name: "first page", query: "?limit=2&offset=0", want: orders[:2]},
		{name: "second order", query: "?limit=1&offset=1", want: orders[1:2]},
		{name: "partial last page", query: "?limit=2&offset=20", want: orders[20:]},
		{name: "maximum limit", query: "?limit=100", want: orders},
		{name: "default limit with offset", query: "?offset=1", want: orders[1:]},
		{name: "offset equals count", query: "?offset=21", want: []OrderResponse{}},
		{name: "maximum offset", query: "?limit=100&offset=" + strconv.Itoa(math.MaxInt), want: []OrderResponse{}},
		{name: "status new", query: "?status=new", want: orders[:20]},
		{name: "status new with limit and offset", query: "?limit=10&offset=5&status=new", want: orders[5:15]},
		{name: "status cancelled", query: "?status=cancelled", want: []OrderResponse{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := handler
			if tt.empty {
				h = newHandler()
			}
			var got []OrderResponse
			assertJSON(t, request(h, http.MethodGet, path+tt.query, "", ""), 200, &got)
			if got == nil {
				t.Fatal("expected a JSON array, got null")
			}
			if !slices.Equal(got, tt.want) {
				t.Errorf("orders=%+v want=%+v", got, tt.want)
			}
		})
	}
}

type listOrdersStub struct {
	Orders
	list func(model.ListOptions) ([]model.Order, error)
}

func (s listOrdersStub) List(options model.ListOptions) ([]model.Order, error) {
	return s.list(options)
}

func TestListOrdersErrors(t *testing.T) {
	tests := []struct {
		name  string
		query string
	}{
		{name: "empty limit", query: "limit="},
		{name: "empty offset", query: "offset="},
		{name: "non numeric limit", query: "limit=abc"},
		{name: "non numeric offset", query: "offset=abc"},
		{name: "negative limit", query: "limit=-1"},
		{name: "negative offset", query: "offset=-1"},
		{name: "zero limit", query: "limit=0"},
		{name: "limit above maximum", query: "limit=101"},
		{name: "overflowing offset", query: "offset=999999999999999999999999999999"},
		{name: "repeated limit", query: "limit=1&limit=2"},
		{name: "repeated offset", query: "offset=1&offset=2"},
		{name: "unknown parameter", query: "page=1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := http.StatusBadRequest
			want := ErrorResponse{Code: "INVALID_QUERY", Error: "invalid query"}

			orders := listOrdersStub{list: func(model.ListOptions) ([]model.Order, error) {
				return nil, nil
			}}

			handler := New(orders, slog.New(slog.NewTextHandler(io.Discard, nil))).Handler()
			w := request(handler, http.MethodGet, "/api/v1/orders?"+tt.query, "", "")

			var got ErrorResponse
			assertJSON(t, w, status, &got)
			if got != want {
				t.Errorf("error=%+v want=%+v", got, want)
			}
		})
	}
}

func assertJSON(t *testing.T, w *httptest.ResponseRecorder, status int, dst any) {
	t.Helper()
	if w.Code != status {
		t.Fatalf("status=%d want=%d body=%s", w.Code, status, w.Body)
	}
	if got := w.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type=%q", got)
	}
	if err := json.Unmarshal(w.Body.Bytes(), dst); err != nil {
		t.Fatalf("decode JSON: %v; body=%s", err, w.Body)
	}
}

func TestCreateAndReadOrder(t *testing.T) {
	handler := newHandler()
	created := request(handler, "POST", "/api/v1/orders", `{"customer":"  Алиса  "}`, "application/json")
	if created.Code != 201 {
		t.Fatalf("create status=%d body=%s", created.Code, created.Body)
	}
	if got := created.Header().Get("Location"); got != "/api/v1/orders/1" {
		t.Fatalf("Location=%q", got)
	}
	var order OrderResponse
	if err := json.Unmarshal(created.Body.Bytes(), &order); err != nil {
		t.Fatal(err)
	}
	if order.ID != 1 || order.Customer != "Алиса" || order.Status != "new" {
		t.Fatalf("unexpected order: %+v", order)
	}
	read := request(handler, "GET", created.Header().Get("Location"), "", "")
	if read.Code != 200 || read.Body.String() != created.Body.String() {
		t.Fatalf("read status=%d body=%s", read.Code, read.Body)
	}
	// Повторный заказ одного клиента разрешён: имя не является уникальным ключом.
	again := request(handler, "POST", "/api/v1/orders", `{"customer":"Алиса"}`, "application/json")
	if again.Code != 201 || again.Header().Get("Location") != "/api/v1/orders/2" {
		t.Fatalf("second create status=%d Location=%q", again.Code, again.Header().Get("Location"))
	}
}

func TestCancelOrder(t *testing.T) {
	handler := newHandler()
	created := request(handler, http.MethodPost, "/api/v1/orders", `{"customer":"Алиса"}`, "application/json")
	var want OrderResponse
	assertJSON(t, created, 201, &want)
	path := created.Header().Get("Location") + "/status"

	t.Run("invalid target status", func(t *testing.T) {
		var out ErrorResponse
		w := request(handler, http.MethodPatch, path, `{"status":"new"}`, "application/json")
		assertJSON(t, w, 400, &out)
	})

	const body = `{"status":"cancelled"}`
	var got OrderResponse
	w := request(handler, http.MethodPatch, path, body, "application/json")
	assertJSON(t, w, 200, &got)
	want.Status = "cancelled"
	if got != want {
		t.Fatalf("order=%+v want=%+v", got, want)
	}

	var out ErrorResponse
	w = request(handler, http.MethodPatch, path, body, "application/json")
	assertJSON(t, w, 409, &out)
	if out.Code != "ORDER_ALREADY_CANCELLED" {
		t.Errorf("error code=%q want ORDER_ALREADY_CANCELLED", out.Code)
	}
}

func TestErrorsKeepJSONContract(t *testing.T) {
	tests := []struct {
		name, method, path, body, contentType string
		status                                int
		code                                  string
	}{
		{"invalid id", "GET", "/api/v1/orders/abc", "", "", 400, "INVALID_ID"},
		{"zero id", "GET", "/api/v1/orders/0", "", "", 400, "INVALID_ID"},
		{"missing order", "GET", "/api/v1/orders/123", "", "", 404, "ORDER_NOT_FOUND"},
		{"cancel with invalid id", "PATCH", "/api/v1/orders/abc/status", `{"status":"cancelled"}`, "application/json", 400, "INVALID_ID"},
		{"cancel missing order", "PATCH", "/api/v1/orders/123/status", `{"status":"cancelled"}`, "application/json", 404, "ORDER_NOT_FOUND"},
		{"missing route", "GET", "/api/v1/missing", "", "", 404, "ROUTE_NOT_FOUND"},
		{"wrong method", "PUT", "/api/v1/orders/1", "", "", 405, "METHOD_NOT_ALLOWED"},
		{"empty customer", "POST", "/api/v1/orders", `{"customer":" "}`, "application/json", 400, "VALIDATION_ERROR"},
		{"null body", "POST", "/api/v1/orders", `null`, "application/json", 400, "INVALID_JSON"},
		{"wrong type", "POST", "/api/v1/orders", `{"customer":42}`, "application/json", 400, "INVALID_JSON"},
		{"unknown field", "POST", "/api/v1/orders", `{"customer":"Алиса","status":"paid"}`, "application/json", 400, "INVALID_JSON"},
		{"second JSON", "POST", "/api/v1/orders", `{"customer":"Алиса"} {}`, "application/json", 400, "INVALID_JSON"},
		{"wrong media type", "POST", "/api/v1/orders", `{"customer":"Алиса"}`, "text/plain", 415, "UNSUPPORTED_MEDIA_TYPE"},
		{"large body", "POST", "/api/v1/orders", `{"customer":"` + strings.Repeat("a", 1100) + `"}`, "application/json", 413, "BODY_TOO_LARGE"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := request(newHandler(), tt.method, tt.path, tt.body, tt.contentType)
			if w.Code != tt.status {
				t.Fatalf("status=%d want=%d body=%s", w.Code, tt.status, w.Body)
			}
			if got := w.Header().Get("Content-Type"); got != "application/json" {
				t.Fatalf("Content-Type=%q", got)
			}
			var out ErrorResponse
			if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
				t.Fatal(err)
			}
			if out.Code != tt.code || out.Error == "" {
				t.Fatalf("unexpected error: %+v", out)
			}
		})
	}
}

func TestUnsupportedMethodDescribesAllowedMethods(t *testing.T) {
	handler := newHandler()
	for _, tt := range []struct{ method, path, allow string }{
		{"DELETE", "/api/v1/orders/1", "GET, HEAD"},
		{"POST", "/api/v1/status", "GET, HEAD"},
	} {
		w := request(handler, tt.method, tt.path, "", "")
		if w.Code != 405 || w.Header().Get("Allow") != tt.allow {
			t.Errorf("%s %s: status=%d Allow=%q", tt.method, tt.path, w.Code, w.Header().Get("Allow"))
		}
	}
}

func TestStatus(t *testing.T) {
	w := request(newHandler(), "GET", "/api/v1/status", "", "")
	if w.Code != 200 || strings.TrimSpace(w.Body.String()) != `{"status":"ok"}` {
		t.Fatalf("status=%d body=%s", w.Code, w.Body)
	}
}
