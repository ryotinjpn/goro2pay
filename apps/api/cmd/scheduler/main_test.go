package main

import (
	"context"
	"errors"
	"testing"

	"github.com/ryotinjpn/goro2pay/apps/api/internal/wallet"
)

type mockWalletSvc struct {
	resetAllFn func(ctx context.Context) (*wallet.ResetResult, error)
}

func (m *mockWalletSvc) GetBalance(ctx context.Context, userID string) (*wallet.WalletSnapshot, error) {
	return nil, nil
}
func (m *mockWalletSvc) SetBudget(ctx context.Context, userID string, monthlyBudget int) error {
	return nil
}
func (m *mockWalletSvc) Deduct(ctx context.Context, userID string, amount int, key string) (*wallet.DeductResult, error) {
	return nil, nil
}
func (m *mockWalletSvc) ResetAll(ctx context.Context) (*wallet.ResetResult, error) {
	return m.resetAllFn(ctx)
}

func TestRunReset_Success(t *testing.T) {
	called := false
	svc := &mockWalletSvc{
		resetAllFn: func(ctx context.Context) (*wallet.ResetResult, error) {
			called = true
			return &wallet.ResetResult{ProcessedUsers: 5}, nil
		},
	}
	if err := runReset(context.Background(), svc); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !called {
		t.Fatal("expected ResetAll to be called")
	}
}

func TestRunReset_PropagatesError(t *testing.T) {
	wantErr := errors.New("ddb failure")
	svc := &mockWalletSvc{
		resetAllFn: func(ctx context.Context) (*wallet.ResetResult, error) {
			return nil, wantErr
		},
	}
	err := runReset(context.Background(), svc)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected %v, got %v", wantErr, err)
	}
}
