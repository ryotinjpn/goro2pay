// Package delivery は外部デリバリー手配のアダプタ層を提供する。
//
// MVP では実外部 API を呼ばず MockDeliveryAdapter (noop) を採用する
// (Application Design Q-F=A、unit-of-work.md §3.3)。本番化時に実 Adapter
// を追加可能な interface 設計とする。
package delivery

import (
	"context"
	"log/slog"
)

// PlaceOrderRequest は DeliveryAdapter.Place に渡すリクエスト。
//
// 凍結契約 §6 では公開 DTO 定義のみ凍結 (内部属性は Unit C 内自由)。
type PlaceOrderRequest struct {
	OrderID        string
	UserID         string
	StoreName      string
	MenuName       string
	Amount         int
	Category       string
	IdempotencyKey string
}

// DeliveryAdapter は外部デリバリー手配の interface。
//
// Place は冪等性キーを引数に取り、同一 key に対して 2 回目以降も冪等な
// レスポンスを返すべき (BR-C09 準拠)。MVP の MockDeliveryAdapter は
// noop で nil を返すため冪等性は自然に成立する。
type DeliveryAdapter interface {
	Place(ctx context.Context, req PlaceOrderRequest) error
}

// MockDeliveryAdapter は MVP デフォルト実装で、ログ出力のみで成功扱いする。
//
// 失敗時の補償トランザクション (BR-C09) は将来実装、本実装では成功固定。
type MockDeliveryAdapter struct{}

// NewMockDeliveryAdapter は MockDeliveryAdapter を返す。
func NewMockDeliveryAdapter() *MockDeliveryAdapter {
	return &MockDeliveryAdapter{}
}

// Place は noop で成功を返す。slog にログを残し、本番モニタリングで
// "mock delivery placed" を観測できるようにする。
func (a *MockDeliveryAdapter) Place(ctx context.Context, req PlaceOrderRequest) error {
	slog.InfoContext(ctx, "mock_delivery_placed",
		"orderId", req.OrderID,
		"storeName", req.StoreName,
		"menuName", req.MenuName,
		"amount", req.Amount,
	)
	return nil
}
