package httpapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
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

func TestCreateOrder(t *testing.T) {
	tests := []struct {
		name         string
		body         string
		wantCustomer string
		wantStatus   int
	}{
		{
			name:         "valid customer",
			body:         `{"customer":"Анна"}`,
			wantCustomer: "Анна",
			wantStatus:   http.StatusCreated,
		},
		{
			name:         "valid customer with spaces",
			body:         `{"customer":"   Анна  "}`,
			wantCustomer: "Анна",
			wantStatus:   http.StatusCreated,
		},
		{
			name:       "empty body",
			body:       ``,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "empty customer",
			body:       `{"customer":"   "}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "broken json",
			body:       `{"customer":`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "wrong customer type",
			body:       `{"customer":123}`,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := repository.NewStore()
			handler := httpapi.NewRouter(store)
			request := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(tt.body))
			request.Header.Set("Content-Type", "application/json")

			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)

			if response.Code != tt.wantStatus {
				t.Fatalf("status=%d, want %d, body=%s",
					response.Code, tt.wantStatus, response.Body.String())
			}
			if got := response.Header().Get("Content-Type"); got != "application/json" {
				t.Errorf("Content-Type=%q, want \"application/json\"", got)
			}

			if tt.wantStatus == http.StatusCreated {
				var order repository.Order
				if err := json.Unmarshal(response.Body.Bytes(), &order); err != nil {
					t.Fatal(err)
				}

				if order.ID <= 0 {
					t.Errorf("id=%d, want positive", order.ID)
				}
				if order.Customer != tt.wantCustomer {
					t.Errorf("customer=%q, want %q", order.Customer, tt.wantCustomer)
				}
				if order.Status != "new" {
					t.Errorf("status=%q, want \"new\"", order.Status)
				}

				saved, ok := store.ByID(order.ID)
				if !ok {
					t.Fatal("order was not saved")
				}
				if saved != order {
					t.Errorf("saved order=%v, response=%v", saved, order)
				}
			} else {
				var body struct {
					Error string `json:"error"`
				}
				if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
					t.Fatalf("decode error response: %v", err)
				}
				if body.Error == "" {
					t.Error("expected non empty error message")
				}

				if _, ok := store.ByID(1); ok {
					t.Error("order was created but input is invalid")
				}
			}
		})
	}

}

func TestGetCreatedOrder(t *testing.T) {
	store := repository.NewStore()
	handler := httpapi.NewRouter(store)

	createRequest := httptest.NewRequest(http.MethodPost, "/orders",
		strings.NewReader(`{"customer":"Анна"}`))

	createRequest.Header.Set("Content-Type", "application/json")

	createResponse := httptest.NewRecorder()
	handler.ServeHTTP(createResponse, createRequest)

	if createResponse.Code != http.StatusCreated {
		t.Fatalf("create status=%d, want 201; body=%s", createResponse.Code, createResponse.Body.String())
	}

	var created repository.Order
	if err := json.Unmarshal(createResponse.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode created order: %v", err)
	}
	if created.ID <= 0 {
		t.Fatalf("created id=%d, want positive", created.ID)
	}

	request := httptest.NewRequest(http.MethodGet, "/orders/"+strconv.Itoa(created.ID), nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("get status=%d, want 200; body=%s", response.Code, response.Body.String())
	}
	if got := response.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type=%q, want application/json", got)
	}

	var order repository.Order
	if err := json.Unmarshal(response.Body.Bytes(), &order); err != nil {
		t.Fatalf("decode order: %v", err)
	}
	if order.ID != created.ID || order.Customer != "Анна" || order.Status != "new" {
		t.Errorf("order=%+v, want id=%d, customer=Анна, status=new", order, created.ID)
	}
}

func TestGetOrderErrors(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		wantStatus int
	}{
		{name: "unknown order", id: "42", wantStatus: http.StatusNotFound},
		{name: "non numeric id", id: "abc", wantStatus: http.StatusBadRequest},
		{name: "zero id", id: "0", wantStatus: http.StatusBadRequest},
		{name: "negative id", id: "-1", wantStatus: http.StatusBadRequest},
		{name: "fractional id", id: "1.5", wantStatus: http.StatusBadRequest},
		{name: "overflowing id", id: "999999999999999999999999999999", wantStatus: http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := repository.NewStore()
			handler := httpapi.NewRouter(store)

			request := httptest.NewRequest(http.MethodGet, "/orders/"+tt.id, nil)
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, request)

			if response.Code != tt.wantStatus {
				t.Fatalf("status=%d, want %d; body=%s", response.Code, tt.wantStatus, response.Body.String())
			}
			if got := response.Header().Get("Content-Type"); got != "application/json" {
				t.Errorf("Content-Type=%q, want \"application/json\"", got)
			}
			var body struct {
				Error string `json:"error"`
			}
			if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
				t.Fatalf("decode error response: %v", err)
			}
			if body.Error == "" {
				t.Error("expected non empty error message")
			}
		})
	}
}
