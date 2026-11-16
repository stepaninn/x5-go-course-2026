package service_test

import (
	"errors"
	"testing"

	"example.com/delivery/model"
	"example.com/delivery/repository"
	"example.com/delivery/service"
)

func TestCalculatorCalculate(t *testing.T) {
	tests := []struct {
		name    string
		order   model.Order
		want    int64
		wantErr error
	}{
		{
			name: "courier",
			order: model.Order{
				ID:          "order-1",
				Tariff:      "courier",
				WeightGrams: 2_500,
				DistanceKM:  10,
			},
			want: 50_000,
		},
		{
			name: "pickup",
			order: model.Order{
				ID:          "order-2",
				Tariff:      "pickup",
				WeightGrams: 3_200,
				DistanceKM:  42,
			},
			want: 39_000,
		},
		{
			name: "express",
			order: model.Order{
				ID:          "order-3",
				Tariff:      "express",
				WeightGrams: 27_000,
				DistanceKM:  210,
			},
			want: 950_000,
		},
		{
			name: "unknown tariff",
			order: model.Order{
				ID:          "order-4",
				Tariff:      "drone",
				WeightGrams: 1_000,
				DistanceKM:  3,
			},
			wantErr: service.ErrUnknownTariff,
		},
		{
			name: "invalid weight",
			order: model.Order{
				ID:          "order-5",
				Tariff:      "courier",
				WeightGrams: 0,
				DistanceKM:  10,
			},
			wantErr: service.ErrInvalidWeight,
		},
		{
			name: "invalid distance",
			order: model.Order{
				ID:          "order-6",
				Tariff:      "express",
				WeightGrams: 1_000,
				DistanceKM:  0,
			},
			wantErr: service.ErrInvalidDistance,
		},
		{
			name: "distance too large",
			order: model.Order{
				ID:          "order-7",
				Tariff:      "pickup",
				WeightGrams: 1_000,
				DistanceKM:  101,
			},
			wantErr: service.ErrDistanceTooLarge,
		},
		{
			name: "courier minimum weight and distance",
			order: model.Order{
				ID:          "order-8",
				Tariff:      "courier",
				WeightGrams: 1,
				DistanceKM:  1,
			},
			want: 26_500,
		},
		{
			name: "courier exact kilogram",
			order: model.Order{
				ID:          "order-9",
				Tariff:      "courier",
				WeightGrams: 1_000,
				DistanceKM:  1,
			},
			want: 26_500,
		},
		{
			name: "courier above kilogram",
			order: model.Order{
				ID:          "order-10",
				Tariff:      "courier",
				WeightGrams: 1_001,
				DistanceKM:  1,
			},
			want: 31_500,
		},
		{
			name: "negative weight",
			order: model.Order{
				ID:          "order-11",
				Tariff:      "courier",
				WeightGrams: -1,
				DistanceKM:  10,
			},
			wantErr: service.ErrInvalidWeight,
		},
		{
			name: "negative distance",
			order: model.Order{
				ID:          "order-12",
				Tariff:      "courier",
				WeightGrams: 1_000,
				DistanceKM:  -1,
			},
			wantErr: service.ErrInvalidDistance,
		},
		{
			name: "courier maximum distance",
			order: model.Order{
				ID:          "order-13",
				Tariff:      "courier",
				WeightGrams: 1_000,
				DistanceKM:  100,
			},
			want: 175_000,
		},
		{
			name: "courier above maximum distance",
			order: model.Order{
				ID:          "order-14",
				Tariff:      "courier",
				WeightGrams: 1_000,
				DistanceKM:  101,
			},
			wantErr: service.ErrDistanceTooLarge,
		},
		{
			name: "pickup maximum distance",
			order: model.Order{
				ID:          "order-15",
				Tariff:      "pickup",
				WeightGrams: 1_000,
				DistanceKM:  50,
			},
			want: 37_000,
		},
		{
			name: "pickup above maximum distance",
			order: model.Order{
				ID:          "order-16",
				Tariff:      "pickup",
				WeightGrams: 1_000,
				DistanceKM:  51,
			},
			wantErr: service.ErrDistanceTooLarge,
		},
		{
			name: "express maximum distance",
			order: model.Order{
				ID:          "order-17",
				Tariff:      "express",
				WeightGrams: 1_000,
				DistanceKM:  300,
			},
			want: 960_000,
		},
		{
			name: "express above maximum distance",
			order: model.Order{
				ID:          "order-18",
				Tariff:      "express",
				WeightGrams: 1_000,
				DistanceKM:  301,
			},
			wantErr: service.ErrDistanceTooLarge,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := repository.Default()
			calculator := service.NewCalculator(repo)

			got, err := calculator.Calculate(tt.order)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("errors.Is(%v, %v) = false", err, tt.wantErr)
			}

			if got != tt.want {
				t.Fatalf("Calculate() got = %d, want = %d", got, tt.want)
			}

			if tt.wantErr != nil {
				var priceErr *service.OrderPriceError
				if !errors.As(err, &priceErr) {
					t.Fatalf("errors.As(%v, *OrderPriceError) = false", err)
				}
				if priceErr.OrderID != tt.order.ID ||
					priceErr.Tariff != tt.order.Tariff {
					t.Errorf("incorrect error context: order ID = %s, got = %s; order tarrif = %s, got = %s",
						tt.order.ID, priceErr.OrderID, tt.order.Tariff, priceErr.Tariff)
				}
			}
		})
	}
}
