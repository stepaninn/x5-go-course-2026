package service

import (
	"context"
	"fmt"
	"math"
	"slices"
	"strings"
	"time"

	"example.com/order-service/internal/order"
)

type OrderService struct {
	repository order.Repository
}

func NewOrderService(repository order.Repository) *OrderService {
	return &OrderService{repository: repository}
}

func validateItem(item order.Item) error {
	if strings.TrimSpace(item.Name) == "" {
		return fmt.Errorf("empty name: %w", order.ErrInvalidOrder)
	}
	if item.Quantity <= 0 {
		return fmt.Errorf("quantity %d: %w", item.Quantity, order.ErrInvalidOrder)
	}
	if item.Price <= 0 {
		return fmt.Errorf("price %d: %w", item.Price, order.ErrInvalidOrder)
	}
	return nil
}

func (s *OrderService) Create(ctx context.Context, input order.CreateOrderInput) (order.Order, error) {
	if len(input.Items) == 0 {
		return order.Order{}, fmt.Errorf("no items: %w", order.ErrInvalidOrder)
	}

	var total int64
	for i, it := range input.Items {
		if err := validateItem(it); err != nil {
			return order.Order{}, fmt.Errorf("item %d: %w", i, err)
		}

		quantity := int64(it.Quantity)
		if it.Price > (math.MaxInt64-total)/quantity {
			return order.Order{}, fmt.Errorf("item %d, total price overflow: %w", i, order.ErrInvalidOrder)
		}
		total += int64(it.Quantity) * it.Price
	}

	now := time.Now().UTC()

	return s.repository.Create(ctx, order.Order{
		Items:      input.Items,
		TotalPrice: total,
		Status:     order.StatusNew,
		CreatedAt:  now,
		History: []order.StatusChange{
			{Status: order.StatusNew, ChangedAt: now},
		},
	})
}

func (s *OrderService) GetByID(ctx context.Context, id int64) (order.Order, error) {
	o, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return order.Order{}, fmt.Errorf("order id %d: %w", id, err)
	}

	return o, nil
}

func (s *OrderService) List(ctx context.Context) ([]order.Order, error) {
	return s.repository.List(ctx)
}

func (s *OrderService) UpdateStatus(ctx context.Context, id int64, status order.Status) (order.Order, error) {
	o, err := s.GetByID(ctx, id)
	if err != nil {
		return order.Order{}, err
	}

	if !status.Valid() {
		return order.Order{}, fmt.Errorf("status %q: %w", status, order.ErrInvalidStatus)
	}

	allowed := (o.Status == order.StatusNew && (status == order.StatusProcessing || status == order.StatusCancelled)) ||
		(o.Status == order.StatusProcessing && (status == order.StatusCancelled || status == order.StatusCompleted))
	if !allowed {
		return order.Order{}, fmt.Errorf(
			"order id %d, transition %q --> %q: %w", id, o.Status, status, order.ErrInvalidStatusTransition)
	}

	o.Status = status
	o.History = append(o.History, order.StatusChange{Status: status, ChangedAt: time.Now().UTC()})

	return s.repository.Update(ctx, o)
}

func (s *OrderService) History(ctx context.Context, id int64) ([]order.StatusChange, error) {
	o, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("order id %d: %w", id, err)
	}

	history := o.History
	slices.SortStableFunc(history, func(a, b order.StatusChange) int {
		return a.ChangedAt.Compare(b.ChangedAt)
	})

	return history, nil
}
