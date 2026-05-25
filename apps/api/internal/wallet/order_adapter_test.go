package wallet

import (
	"context"
	"errors"
	"testing"

	"github.com/ryotinjpn/goro2pay/apps/api/internal/order"
)

func TestAdapter_Deduct_Success(t *testing.T) {
	svc := &mockSvc{
		deductFn: func(ctx context.Context, userID string, amount int, key string) (*DeductResult, error) {
			if amount != 850 || key != "user_a:01HW3XKQ7EZ8XN1MX1C7Q4Z9TY" {
				t.Fatalf("unexpected args: amount=%d key=%s", amount, key)
			}
			return &DeductResult{NewBalance: 29150, Idempotent: false}, nil
		},
	}
	ad := NewOrderAdapter(svc)
	res, err := ad.Deduct(context.Background(), "user_a", "user_a:01HW3XKQ7EZ8XN1MX1C7Q4Z9TY", 850)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if res.RemainingBalance != 29150 || res.Idempotent {
		t.Fatalf("unexpected result: %+v", res)
	}
	if res.OrderID != "user_a:01HW3XKQ7EZ8XN1MX1C7Q4Z9TY" {
		t.Fatalf("expected OrderID == idempotencyKey, got %s", res.OrderID)
	}
}

func TestAdapter_Deduct_InsufficientBalance(t *testing.T) {
	svc := &mockSvc{
		deductFn: func(ctx context.Context, userID string, amount int, key string) (*DeductResult, error) {
			return nil, ErrInsufficientBalance
		},
	}
	ad := NewOrderAdapter(svc)
	_, err := ad.Deduct(context.Background(), "user_a", "user_a:01HW3XKQ7EZ8XN1MX1C7Q4Z9TY", 5000)
	if !errors.Is(err, order.ErrInsufficientFunds) {
		t.Fatalf("expected order.ErrInsufficientFunds, got %v", err)
	}
}

func TestAdapter_Deduct_IdempotencyConflict(t *testing.T) {
	svc := &mockSvc{
		deductFn: func(ctx context.Context, userID string, amount int, key string) (*DeductResult, error) {
			return nil, ErrIdempotencyConflict
		},
	}
	ad := NewOrderAdapter(svc)
	_, err := ad.Deduct(context.Background(), "user_a", "user_a:01HW3XKQ7EZ8XN1MX1C7Q4Z9TY", 850)
	if !errors.Is(err, order.ErrIdempotencyConflict) {
		t.Fatalf("expected order.ErrIdempotencyConflict, got %v", err)
	}
}

func TestAdapter_Deduct_Idempotent(t *testing.T) {
	svc := &mockSvc{
		deductFn: func(ctx context.Context, userID string, amount int, key string) (*DeductResult, error) {
			return &DeductResult{NewBalance: 29150, Idempotent: true}, nil
		},
	}
	ad := NewOrderAdapter(svc)
	res, err := ad.Deduct(context.Background(), "user_a", "user_a:01HW3XKQ7EZ8XN1MX1C7Q4Z9TY", 850)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !res.Idempotent {
		t.Fatal("expected Idempotent=true")
	}
}
