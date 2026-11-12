package service_test

import (
	"context"
	"errors"
	"math"
	"testing"
	"time"

	"example.com/order-service/internal/order"
	"example.com/order-service/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testRepository struct {
	order.Repository
	orders    []order.Order
	listErr   error
	value     order.Order
	getErr    error
	updateErr error
}

func (r *testRepository) GetByID(ctx context.Context, id int64) (order.Order, error) {
	return r.value, r.getErr
}

func (r *testRepository) Update(ctx context.Context, value order.Order) (order.Order, error) {
	if r.updateErr != nil {
		return order.Order{}, r.updateErr
	}
	r.value = value
	return value, nil
}

func (r *testRepository) Create(ctx context.Context, value order.Order) (order.Order, error) {
	return value, nil
}

func (r *testRepository) List(ctx context.Context) ([]order.Order, error) {
	return r.orders, r.listErr
}

func TestListOrders(t *testing.T) {
	repositoryErr := errors.New("repository error")

	tests := []struct {
		name    string
		orders  []order.Order
		wantErr error
	}{
		{
			name: "success 1",
			orders: []order.Order{
				{ID: 1, Status: order.StatusNew, TotalPrice: 100},
			},
		},
		{
			name: "success 2",
			orders: []order.Order{
				{ID: 2, Status: order.StatusProcessing, TotalPrice: 200},
				{ID: 1, Status: order.StatusNew, TotalPrice: 100},
			},
		},
		{
			name:   "empty orders",
			orders: []order.Order{},
		},
		{
			name: "nil orders",
		},
		{
			name:    "repository error",
			wantErr: repositoryErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			svc := service.NewOrderService(&testRepository{
				orders:  tt.orders,
				listErr: tt.wantErr,
			})

			orders, err := svc.List(ctx)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.orders, orders)
		})
	}
}

func TestCreateOrder(t *testing.T) {
	tests := []struct {
		name       string
		input      order.CreateOrderInput
		totalPrice int64
		wantErr    error
	}{
		{
			name: "success 1",
			input: order.CreateOrderInput{Items: []order.Item{
				{
					Name:     "test item 1",
					Quantity: 1,
					Price:    1,
				},
				{
					Name:     "test item 2",
					Quantity: 2,
					Price:    2,
				},
			},
			},
			totalPrice: 1*1 + 2*2,
			wantErr:    nil,
		},
		{
			name: "success 2",
			input: order.CreateOrderInput{Items: []order.Item{
				{
					Name:     "test item 1",
					Quantity: 12,
					Price:    23,
				},
				{
					Name:     "test item 2",
					Quantity: 21,
					Price:    32,
				},
			},
			},
			totalPrice: 12*23 + 21*32,
			wantErr:    nil,
		},
		{
			name: "success 3",
			input: order.CreateOrderInput{Items: []order.Item{
				{
					Name:     "test item 1",
					Quantity: 4522,
					Price:    73,
				},
				{
					Name:     "test item 2",
					Quantity: 3963,
					Price:    46,
				},
			},
			},
			totalPrice: 4522*73 + 3963*46,
			wantErr:    nil,
		},
		{
			name: "total at int64 limit",
			input: order.CreateOrderInput{Items: []order.Item{
				{Name: "first", Quantity: 1, Price: math.MaxInt64 - 1},
				{Name: "second", Quantity: 1, Price: 1},
			}},
			totalPrice: math.MaxInt64,
		},
		{
			name: "mul overflow",
			input: order.CreateOrderInput{Items: []order.Item{
				{Name: "item", Quantity: 2, Price: math.MaxInt64/2 + 1},
			}},
			wantErr: order.ErrInvalidOrder,
		},
		{
			name: "sum overflow",
			input: order.CreateOrderInput{Items: []order.Item{
				{Name: "first", Quantity: 1, Price: math.MaxInt64},
				{Name: "second", Quantity: 1, Price: 1},
			}},
			wantErr: order.ErrInvalidOrder,
		},
		{
			name:    "order with no items",
			input:   order.CreateOrderInput{Items: []order.Item{}},
			wantErr: order.ErrInvalidOrder,
		},
		{
			name:    "order with nil items",
			input:   order.CreateOrderInput{},
			wantErr: order.ErrInvalidOrder,
		},
		{
			name: "empty item name 1",
			input: order.CreateOrderInput{Items: []order.Item{
				{
					Name:     "test item 1",
					Quantity: 1,
					Price:    1,
				},
				{
					Name:     "",
					Quantity: 2,
					Price:    2,
				},
			},
			},
			wantErr: order.ErrInvalidOrder,
		},
		{
			name: "empty item name 2",
			input: order.CreateOrderInput{Items: []order.Item{
				{Name: " \t\n", Quantity: 1, Price: 1},
			}},
			wantErr: order.ErrInvalidOrder,
		},
		{
			name: "zero quantity",
			input: order.CreateOrderInput{Items: []order.Item{
				{
					Name:     "test item 1",
					Quantity: 1,
					Price:    1,
				},
				{
					Name:     "test item 2",
					Quantity: 0,
					Price:    2,
				},
			},
			},
			wantErr: order.ErrInvalidOrder,
		},
		{
			name: "negative quantity",
			input: order.CreateOrderInput{Items: []order.Item{
				{
					Name:     "test item 1",
					Quantity: -1,
					Price:    1,
				},
				{
					Name:     "test item 2",
					Quantity: 2,
					Price:    2,
				},
			},
			},
			wantErr: order.ErrInvalidOrder,
		},
		{
			name: "zero price",
			input: order.CreateOrderInput{Items: []order.Item{
				{
					Name:     "test item 1",
					Quantity: 1,
					Price:    0,
				},
				{
					Name:     "test item 2",
					Quantity: 2,
					Price:    2,
				},
			},
			},
			wantErr: order.ErrInvalidOrder,
		},
		{
			name: "negative price",
			input: order.CreateOrderInput{Items: []order.Item{
				{
					Name:     "test item 1",
					Quantity: 1,
					Price:    1,
				},
				{
					Name:     "test item 2",
					Quantity: 2,
					Price:    -2,
				},
			},
			},
			wantErr: order.ErrInvalidOrder,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			svc := service.NewOrderService(&testRepository{})

			o, err := svc.Create(ctx, tt.input)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.input.Items, o.Items)
			assert.Equal(t, order.StatusNew, o.Status)
			assert.Equal(t, tt.totalPrice, o.TotalPrice)

			require.Equal(t, len(o.History), 1)
			assert.Equal(t, order.StatusNew, o.History[0].Status)
			assert.Equal(t, o.CreatedAt, o.History[0].ChangedAt)
		})
	}
}

func TestUpdateStatus(t *testing.T) {
	repositoryErr := errors.New("repository error")

	tests := []struct {
		name      string
		from      order.Status
		to        order.Status
		getErr    error
		updateErr error
		wantErr   error
	}{
		{
			name: "new to processing",
			from: order.StatusNew,
			to:   order.StatusProcessing,
		},
		{
			name: "new to cancelled",
			from: order.StatusNew,
			to:   order.StatusCancelled,
		},
		{
			name: "processing to completed",
			from: order.StatusProcessing,
			to:   order.StatusCompleted,
		},
		{
			name: "processing to cancelled",
			from: order.StatusProcessing,
			to:   order.StatusCancelled,
		},
		{
			name:    "new to new",
			from:    order.StatusNew,
			to:      order.StatusNew,
			wantErr: order.ErrInvalidStatusTransition,
		},
		{
			name:    "new to completed",
			from:    order.StatusNew,
			to:      order.StatusCompleted,
			wantErr: order.ErrInvalidStatusTransition,
		},
		{
			name:    "processing to new",
			from:    order.StatusProcessing,
			to:      order.StatusNew,
			wantErr: order.ErrInvalidStatusTransition,
		},
		{
			name:    "processing to processing",
			from:    order.StatusProcessing,
			to:      order.StatusProcessing,
			wantErr: order.ErrInvalidStatusTransition,
		},
		{
			name:    "completed to new",
			from:    order.StatusCompleted,
			to:      order.StatusNew,
			wantErr: order.ErrInvalidStatusTransition,
		},
		{
			name:    "completed to processing",
			from:    order.StatusCompleted,
			to:      order.StatusProcessing,
			wantErr: order.ErrInvalidStatusTransition,
		},
		{
			name:    "completed to completed",
			from:    order.StatusCompleted,
			to:      order.StatusCompleted,
			wantErr: order.ErrInvalidStatusTransition,
		},
		{
			name:    "completed to cancelled",
			from:    order.StatusCompleted,
			to:      order.StatusCancelled,
			wantErr: order.ErrInvalidStatusTransition,
		},
		{
			name:    "cancelled to new",
			from:    order.StatusCancelled,
			to:      order.StatusNew,
			wantErr: order.ErrInvalidStatusTransition,
		},
		{
			name:    "cancelled to processing",
			from:    order.StatusCancelled,
			to:      order.StatusProcessing,
			wantErr: order.ErrInvalidStatusTransition,
		},
		{
			name:    "cancelled to completed",
			from:    order.StatusCancelled,
			to:      order.StatusCompleted,
			wantErr: order.ErrInvalidStatusTransition,
		},
		{
			name:    "cancelled to cancelled",
			from:    order.StatusCancelled,
			to:      order.StatusCancelled,
			wantErr: order.ErrInvalidStatusTransition,
		},
		{
			name:    "empty status",
			from:    order.StatusNew,
			to:      "",
			wantErr: order.ErrInvalidStatus,
		},
		{
			name:    "unknown status",
			from:    order.StatusNew,
			to:      "unknown",
			wantErr: order.ErrInvalidStatus,
		},
		{
			name:    "order not found",
			to:      order.StatusProcessing,
			getErr:  order.ErrOrderNotFound,
			wantErr: order.ErrOrderNotFound,
		},
		{
			name:    "get error",
			to:      order.StatusProcessing,
			getErr:  repositoryErr,
			wantErr: repositoryErr,
		},
		{
			name:      "update error",
			from:      order.StatusNew,
			to:        order.StatusProcessing,
			updateErr: repositoryErr,
			wantErr:   repositoryErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			firstChange := order.StatusChange{
				Status:    tt.from,
				ChangedAt: time.Date(2026, time.January, 1, 12, 0, 0, 0, time.UTC),
			}
			repo := &testRepository{
				value: order.Order{
					ID:      1,
					Status:  tt.from,
					History: []order.StatusChange{firstChange},
				},
				getErr:    tt.getErr,
				updateErr: tt.updateErr,
			}
			svc := service.NewOrderService(repo)

			o, err := svc.UpdateStatus(ctx, 1, tt.to)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Equal(t, tt.from, repo.value.Status)
				require.Equal(t, len(repo.value.History), 1)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, int64(1), o.ID)
			assert.Equal(t, tt.to, o.Status)
			assert.Equal(t, o, repo.value)
			require.Equal(t, len(o.History), 2)
			assert.Equal(t, firstChange, o.History[0])
			assert.Equal(t, tt.to, o.History[1].Status)
		})
	}
}

func TestGetOrderByID(t *testing.T) {
	tests := []struct {
		name    string
		value   order.Order
		wantErr error
	}{
		{
			name:  "existing order",
			value: order.Order{ID: 1, Status: order.StatusNew, TotalPrice: 100},
		},
		{
			name:    "order not found",
			wantErr: order.ErrOrderNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			svc := service.NewOrderService(&testRepository{
				value:  tt.value,
				getErr: tt.wantErr,
			})

			o, err := svc.GetByID(ctx, 1)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.value, o)
		})
	}
}

func TestOrderHistory(t *testing.T) {
	first := order.StatusChange{
		Status:    order.StatusNew,
		ChangedAt: time.Date(2026, time.January, 1, 12, 0, 0, 0, time.UTC),
	}
	second := order.StatusChange{
		Status:    order.StatusProcessing,
		ChangedAt: time.Date(2026, time.January, 1, 13, 0, 0, 0, time.UTC),
	}
	third := order.StatusChange{
		Status:    order.StatusCompleted,
		ChangedAt: time.Date(2026, time.January, 1, 14, 0, 0, 0, time.UTC),
	}

	tests := []struct {
		name    string
		history []order.StatusChange
		want    []order.StatusChange
		wantErr error
	}{
		{
			name:    "one change",
			history: []order.StatusChange{first},
			want:    []order.StatusChange{first},
		},
		{
			name:    "sorted history",
			history: []order.StatusChange{first, second, third},
			want:    []order.StatusChange{first, second, third},
		},
		{
			name:    "unsorted history",
			history: []order.StatusChange{third, first, second},
			want:    []order.StatusChange{first, second, third},
		},
		{
			name:    "order not found",
			wantErr: order.ErrOrderNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			svc := service.NewOrderService(&testRepository{
				value:  order.Order{ID: 1, History: tt.history},
				getErr: tt.wantErr,
			})

			history, err := svc.History(ctx, 1)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, history)
		})
	}
}
