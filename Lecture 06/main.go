package main

import (
	"errors"
	"fmt"
	"os"
)

type Order struct {
	ID       int
	Customer string
	Status   string
}

var ErrEmptyCustomer = errors.New("имя клиента не задано")
var ErrNotFound = errors.New("заказ не найден")

type Memory struct {
	orders map[int]Order
}

func NewMemory() *Memory {
	return &Memory{orders: make(map[int]Order)}
}

func (m *Memory) Save(order Order) error {
	m.orders[order.ID] = order
	return nil
}

func (m *Memory) ByID(id int) (Order, error) {
	order, ok := m.orders[id]
	if !ok {
		return Order{}, ErrNotFound
	}
	return order, nil
}

type Service struct {
	repo   *Memory
	nextID int
}

func NewService(repo *Memory) *Service {
	return &Service{repo: repo, nextID: 1}
}

func (s *Service) Create(customer string) (Order, error) {
	if customer == "" {
		return Order{}, ErrEmptyCustomer
	}
	order := Order{
		ID:       s.nextID,
		Customer: customer,
		Status:   "new",
	}
	if err := s.repo.Save(order); err != nil {
		return Order{}, err
	}
	s.nextID++
	return order, nil
}

type Handler struct {
	orders *Service
}

func NewHandler(orders *Service) *Handler {
	return &Handler{orders: orders}
}

func (h *Handler) Create(customer string) (Order, error) {
	return h.orders.Create(customer)
}

func main() {
	repo := NewMemory()
	orders := NewService(repo)
	handler := NewHandler(orders)

	address := os.Getenv("HTTP_ADDR")
	if address == "" {
		address = ":8080"
	}
	order, err := handler.Create("Покупатель")
	if err != nil {
		fmt.Println("ошибка:", err)
		return
	}
	fmt.Printf("сервис слушал бы %s, создан заказ: %+v\n", address, order)
}
