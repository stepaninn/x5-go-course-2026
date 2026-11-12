package memory

import (
	"context"
	"slices"
	"sync"

	"example.com/order-service/internal/order"
)

type OrderRepository struct {
	mu sync.Mutex
	// единственное мне кажется, что массив плох для будущего расширения
	// будто лучше использовать мапу, раз на то пошло
	items  []order.Order
	nextID int64
}

func NewOrderRepository() *OrderRepository {
	return &OrderRepository{
		items:  make([]order.Order, 0),
		nextID: 1,
	}
}

// написал глубокую копию, чтобы корректно копировать срезы в Order
func cloneOrder(o order.Order) order.Order {
	cloned := o

	if o.Items != nil {
		cloned.Items = make([]order.Item, len(o.Items))
		copy(cloned.Items, o.Items)
	}
	if o.History != nil {
		cloned.History = make([]order.StatusChange, len(o.History))
		copy(cloned.History, o.History)
	}

	return cloned
}

func (r *OrderRepository) Create(_ context.Context, value order.Order) (order.Order, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	item := cloneOrder(value)

	item.ID = r.nextID
	r.nextID++

	r.items = append(r.items, item)

	return cloneOrder(item), nil
}

func (r *OrderRepository) GetByID(_ context.Context, id int64) (order.Order, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if id <= 0 || id > int64(len(r.items)) {
		return order.Order{}, order.ErrOrderNotFound
	}

	item := r.items[id-1]

	return cloneOrder(item), nil
}

// ну и раз здесь просят без доступа к внутрянке, то cloneOrder тоже заходит
func (r *OrderRepository) List(_ context.Context) ([]order.Order, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	items := make([]order.Order, 0, len(r.items))
	for _, v := range r.items {
		items = append(items, cloneOrder(v))
	}

	// я так понял это подразумевается под временем создания в todo
	slices.SortStableFunc(items, func(a, b order.Order) int {
		return a.CreatedAt.Compare(b.CreatedAt)
	})

	return items, nil
}

func (r *OrderRepository) Update(_ context.Context, value order.Order) (order.Order, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if value.ID <= 0 || value.ID > int64(len(r.items)) {
		return order.Order{}, order.ErrOrderNotFound
	}

	item := cloneOrder(value)
	r.items[item.ID-1] = item

	return cloneOrder(item), nil
}
