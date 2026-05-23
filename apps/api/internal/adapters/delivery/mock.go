package delivery

import "context"

// FakeDeliveryAdapter は DeliveryAdapter の関数フィールド注入型 mock (P-MOCK-01)。
//
// MockDeliveryAdapter は production の noop 実装、FakeDeliveryAdapter は
// テスト時にエラーパス等を制御するための fake (両者は責務が異なる)。
type FakeDeliveryAdapter struct {
	PlaceFunc func(ctx context.Context, req PlaceOrderRequest) error
	Calls     int
}

// Place は FakeDeliveryAdapter の interface 実装。
func (f *FakeDeliveryAdapter) Place(ctx context.Context, req PlaceOrderRequest) error {
	f.Calls++
	if f.PlaceFunc != nil {
		return f.PlaceFunc(ctx, req)
	}
	return nil
}
