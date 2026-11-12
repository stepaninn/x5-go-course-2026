package service_test

import (
	"context"
	"math"
	"testing"

	"example.com/order-service/internal/order"
	"example.com/order-service/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testRepository struct {
	order.Repository
}

func (r *testRepository) Create(ctx context.Context, value order.Order) (order.Order, error) {
	return value, nil
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
