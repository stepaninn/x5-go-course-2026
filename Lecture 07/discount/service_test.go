package discount

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidatePercent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		percent int
		wantErr error
	}{
		{name: "zero percent", percent: 0},
		{name: "ten percent", percent: 10},
		{name: "twenty five percent", percent: 25},
		{name: "fifty percent", percent: 50},
		{name: "hundred percent", percent: 100},
		{name: "negative percent", percent: -1, wantErr: ErrInvalidPercent},
		{name: "percent above hundred", percent: 101, wantErr: ErrInvalidPercent},
		{name: "large positive percent", percent: 10000, wantErr: ErrInvalidPercent},
		{name: "large negative percent", percent: -10000, wantErr: ErrInvalidPercent},
		{name: "maximum int", percent: math.MaxInt, wantErr: ErrInvalidPercent},
		{name: "minimum int", percent: math.MinInt, wantErr: ErrInvalidPercent},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidatePercent(tt.percent)
			if tt.wantErr != nil {
				assert.ErrorIs(t, tt.wantErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestApply(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		price   int
		percent int
		want    int
		wantErr error
	}{
		{name: "regular discount", price: 200, percent: 25, want: 150},
		{name: "zero discount", price: 200, percent: 0, want: 200},
		{name: "full discount", price: 200, percent: 100, want: 0},
		{name: "zero price", price: 0, percent: 10, want: 0},
		{name: "round integer", price: 99, percent: 10, want: 89},
		{name: "negative price", price: -1, percent: 10, wantErr: ErrNegativePrice},
		{name: "negative percent", price: 200, percent: -1, wantErr: ErrInvalidPercent},
		{name: "percent above hundred", price: 200, percent: 101, wantErr: ErrInvalidPercent},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := Apply(tt.price, tt.percent)
			if tt.wantErr != nil {
				assert.ErrorIs(t, tt.wantErr, err)
			} else {
				assert.Equal(t, tt.want, got)
				assert.NoError(t, err)
			}
		})
	}
}

func TestFinalPrice(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		items   []Item
		want    int
		wantErr error
	}{
		{name: "nil list", items: nil, want: 0},
		{name: "empty list", items: []Item{}, want: 0},
		{
			name: "many valid items",
			items: []Item{
				{Price: 200, DiscountPercent: 25},
				{Price: 100, DiscountPercent: 0},
				{Price: 300, DiscountPercent: 100},
			},
			want: 250,
		},
		{
			name: "negative price after valid item",
			items: []Item{
				{Price: 200, DiscountPercent: 25},
				{Price: -1, DiscountPercent: 10},
			},
			wantErr: ErrNegativePrice,
		},
		{
			name:    "negative discount",
			items:   []Item{{Price: 100, DiscountPercent: -1}},
			wantErr: ErrInvalidPercent,
		},
		{
			name: "discount above hundred after valid item",
			items: []Item{
				{Price: 200, DiscountPercent: 25},
				{Price: 100, DiscountPercent: 101},
			},
			wantErr: ErrInvalidPercent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := FinalPrice(tt.items)
			if tt.wantErr != nil {
				require.ErrorIs(t, tt.wantErr, err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
