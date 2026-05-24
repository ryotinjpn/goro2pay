package delivery

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockDeliveryAdapter_Place(t *testing.T) {
	a := NewMockDeliveryAdapter()
	err := a.Place(context.Background(), PlaceOrderRequest{
		OrderID:   "01HZ",
		StoreName: "ゴロゴロ食堂",
	})
	require.NoError(t, err)
}

func TestFakeDeliveryAdapter_DefaultBehavior(t *testing.T) {
	f := &FakeDeliveryAdapter{}
	err := f.Place(context.Background(), PlaceOrderRequest{})
	require.NoError(t, err)
	assert.Equal(t, 1, f.Calls)
}

func TestFakeDeliveryAdapter_CustomFunc(t *testing.T) {
	expectedErr := errors.New("delivery service down")
	f := &FakeDeliveryAdapter{
		PlaceFunc: func(ctx context.Context, req PlaceOrderRequest) error {
			return expectedErr
		},
	}
	err := f.Place(context.Background(), PlaceOrderRequest{})
	assert.ErrorIs(t, err, expectedErr)
}
