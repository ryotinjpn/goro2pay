package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ryotinjpn/goro2pay/apps/api/internal/order"
)

// fakeOrderService は OrderService の手書き mock。
type fakeOrderService struct {
	placeOrderFn func(ctx context.Context, userID string, req order.PlaceOrderRequest) (*order.PlaceOrderResult, error)
	getHistoryFn func(ctx context.Context, userID string, limit int) ([]*order.OrderRecord, error)
}

func (f *fakeOrderService) PlaceOrder(ctx context.Context, userID string, req order.PlaceOrderRequest) (*order.PlaceOrderResult, error) {
	return f.placeOrderFn(ctx, userID, req)
}
func (f *fakeOrderService) GetHistory(ctx context.Context, userID string, limit int) ([]*order.OrderRecord, error) {
	return f.getHistoryFn(ctx, userID, limit)
}

func setupRouter(svc *fakeOrderService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		// auth middleware を skip して userId を直接 set
		// (logging.CtxKeyUserID = "userId" 文字列キー)
		c.Set("userId", "user-1")
		c.Next()
	})
	h := NewOrderHandler(svc)
	r.POST("/api/orders", h.PlaceOrder)
	r.GET("/api/orders", h.GetHistory)
	return r
}

func TestOrderHandler_PlaceOrder_Success(t *testing.T) {
	svc := &fakeOrderService{
		placeOrderFn: func(ctx context.Context, userID string, req order.PlaceOrderRequest) (*order.PlaceOrderResult, error) {
			return &order.PlaceOrderResult{OrderID: "01HZ", StoreName: "ゴロゴロ食堂", MenuName: "おまかせ定食", Amount: 1000, RemainingBalance: 9000}, nil
		},
	}
	r := setupRouter(svc)
	body := `{"category":"food","idempotencyKey":"01HZIDEM"}`
	req := httptest.NewRequest(http.MethodPost, "/api/orders", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var res placeOrderResponseDTO
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &res))
	assert.Equal(t, "01HZ", res.OrderID)
	assert.Equal(t, 9000, res.RemainingBalance)
}

func TestOrderHandler_PlaceOrder_InsufficientFunds_402(t *testing.T) {
	svc := &fakeOrderService{
		placeOrderFn: func(ctx context.Context, userID string, req order.PlaceOrderRequest) (*order.PlaceOrderResult, error) {
			return nil, order.ErrInsufficientFunds
		},
	}
	r := setupRouter(svc)
	body := `{"category":"food","idempotencyKey":"01HZ"}`
	req := httptest.NewRequest(http.MethodPost, "/api/orders", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusPaymentRequired, w.Code)
	var res errorResponseDTO
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &res))
	assert.Equal(t, "INSUFFICIENT_FUNDS", res.Error.Code)
}

func TestOrderHandler_PlaceOrder_IdempotencyConflict_409(t *testing.T) {
	svc := &fakeOrderService{
		placeOrderFn: func(ctx context.Context, userID string, req order.PlaceOrderRequest) (*order.PlaceOrderResult, error) {
			return nil, order.ErrIdempotencyConflict
		},
	}
	r := setupRouter(svc)
	body := `{"category":"food","idempotencyKey":"01HZ"}`
	req := httptest.NewRequest(http.MethodPost, "/api/orders", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestOrderHandler_PlaceOrder_InternalError_500(t *testing.T) {
	svc := &fakeOrderService{
		placeOrderFn: func(ctx context.Context, userID string, req order.PlaceOrderRequest) (*order.PlaceOrderResult, error) {
			return nil, errors.New("downstream failure")
		},
	}
	r := setupRouter(svc)
	body := `{"category":"food","idempotencyKey":"01HZ"}`
	req := httptest.NewRequest(http.MethodPost, "/api/orders", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestOrderHandler_PlaceOrder_BadRequest_400(t *testing.T) {
	svc := &fakeOrderService{
		placeOrderFn: func(ctx context.Context, userID string, req order.PlaceOrderRequest) (*order.PlaceOrderResult, error) {
			return nil, nil
		},
	}
	r := setupRouter(svc)
	body := `{"category":""}`
	req := httptest.NewRequest(http.MethodPost, "/api/orders", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestOrderHandler_GetHistory_Success(t *testing.T) {
	svc := &fakeOrderService{
		getHistoryFn: func(ctx context.Context, userID string, limit int) ([]*order.OrderRecord, error) {
			return []*order.OrderRecord{
				{OrderID: "01HZ", Category: "food", StoreName: "X", MenuName: "Y", Amount: 1000, OrderedAt: "2026-05-24T00:00:00Z"},
			}, nil
		},
	}
	r := setupRouter(svc)
	req := httptest.NewRequest(http.MethodGet, "/api/orders?limit=10", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var res getHistoryResponseDTO
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &res))
	require.Len(t, res.Items, 1)
	assert.Equal(t, "01HZ", res.Items[0].OrderID)
}

func TestParsePositiveInt(t *testing.T) {
	tests := []struct {
		input    string
		expected int
		wantErr  bool
	}{
		{"20", 20, false},
		{"100", 100, false},
		{"0", 0, true},
		{"-5", 0, true},
		{"abc", 0, true},
		{"", 0, true},
		{"99999", 0, true},
	}
	for _, tt := range tests {
		got, err := parsePositiveInt(tt.input)
		if tt.wantErr {
			assert.Error(t, err, "input=%s", tt.input)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, got)
		}
	}
}
