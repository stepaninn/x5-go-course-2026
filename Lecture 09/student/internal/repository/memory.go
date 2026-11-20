package repository

import (
	"cmp"
	"errors"
	"slices"
	"sync"

	"example.com/lesson9practice/internal/model"
)

var ErrNotFound = errors.New("order not found")
var ErrAlreadyCancelled = errors.New("order already cancelled")
var ErrInvalidStatusTransition = errors.New("unsupported status change")

type Memory struct {
	mu     sync.Mutex
	nextID int
	orders map[int]model.Order
}

func New() *Memory { return &Memory{nextID: 1, orders: make(map[int]model.Order)} }

func (m *Memory) Create(customer string) (model.Order, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	order := model.Order{ID: m.nextID, Customer: customer, Status: "new"}
	m.orders[order.ID] = order
	m.nextID++
	return order, nil
}
func (m *Memory) ByID(id int) (model.Order, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	order, ok := m.orders[id]
	if !ok {
		return model.Order{}, ErrNotFound
	}
	return order, nil
}

func (m *Memory) ChangeStatusByID(id int, status string) (model.Order, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	order, ok := m.orders[id]
	if !ok {
		return model.Order{}, ErrNotFound
	}

	if order.Status == "cancelled" {
		return model.Order{}, ErrAlreadyCancelled
	}

	if order.Status == status {
		return model.Order{}, ErrInvalidStatusTransition
	}

	if order.Status == "new" && status == "cancelled" {
		order.Status = "cancelled"
		m.orders[id] = order
		return order, nil
	}

	return model.Order{}, ErrInvalidStatusTransition
}

func (m *Memory) List(options model.ListOptions) ([]model.Order, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	orders := make([]model.Order, 0)
	for _, v := range m.orders {
		if options.Status == "" || v.Status == options.Status {
			orders = append(orders, v)
		}

	}

	slices.SortFunc(orders, func(a, b model.Order) int {
		return cmp.Compare(a.ID, b.ID)
	})

	if options.Offset >= len(orders) {
		return []model.Order{}, nil
	}
	orders = orders[options.Offset:]

	if options.Limit >= len(orders) {
		return orders, nil
	}

	return orders[:options.Limit], nil
}
