package metrics

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	"github.com/ryotinjpn/goro2pay/apps/api/internal/repo/budget_settings"
	orderhistory "github.com/ryotinjpn/goro2pay/apps/api/internal/repo/order_history"
	"github.com/ryotinjpn/goro2pay/apps/api/internal/repo/wallet_repo"
)

// --- mock implementations ---

type mockBudgetSettingsReader struct {
	record *budget_settings.BudgetSettings
	err    error
}

func (m *mockBudgetSettingsReader) Get(_ context.Context, _ string) (*budget_settings.BudgetSettings, error) {
	return m.record, m.err
}

type mockWalletReader struct {
	record *wallet_repo.WalletRecord
	err    error
}

func (m *mockWalletReader) Get(_ context.Context, _ string) (*wallet_repo.WalletRecord, error) {
	return m.record, m.err
}

type mockOrderHistoryReader struct {
	count int
	err   error
}

func (m *mockOrderHistoryReader) ListRecent(_ context.Context, _ string, _ int) ([]orderhistory.OrderRecord, error) {
	return nil, nil
}
func (m *mockOrderHistoryReader) CountThisMonth(_ context.Context, _ string) (int, error) {
	return m.count, m.err
}
func (m *mockOrderHistoryReader) SumThisMonth(_ context.Context, _ string) (int, error) {
	return 0, nil
}

// --- table tests ---

func TestGetMetrics(t *testing.T) {
	fixedBS := &budget_settings.BudgetSettings{UserID: "u1", MonthlyBudget: 30000, EffectiveFrom: time.Now()}
	fixedW := &wallet_repo.WalletRecord{UserID: "u1", Balance: 5400}

	cases := []struct {
		name      string
		bs        *budget_settings.BudgetSettings
		bsErr     error
		wallet    *wallet_repo.WalletRecord
		walletErr error
		count     int
		countErr  error
		wantErr   error
		check     func(t *testing.T, m *Metrics)
	}{
		{
			name:   "正常: ThresholdExceeded=true",
			bs:     fixedBS,
			wallet: fixedW,
			count:  12,
			check: func(t *testing.T, m *Metrics) {
				if m.DamageCount != 12 {
					t.Errorf("DamageCount want 12 got %d", m.DamageCount)
				}
				if !m.ThresholdExceeded {
					t.Error("ThresholdExceeded want true")
				}
				if m.RemainingBalance != 5400 {
					t.Errorf("RemainingBalance want 5400 got %d", m.RemainingBalance)
				}
			},
		},
		{
			name:    "BudgetSettings nil → ErrNoBudgetSet",
			bs:      nil,
			wallet:  fixedW,
			wantErr: ErrNoBudgetSet,
		},
		{
			name:    "MonthlyBudget=0 → ErrNoBudgetSet",
			bs:      &budget_settings.BudgetSettings{MonthlyBudget: 0},
			wallet:  fixedW,
			wantErr: ErrNoBudgetSet,
		},
		{
			name:    "Wallet nil → ErrNoBudgetSet",
			bs:      fixedBS,
			wallet:  nil,
			wantErr: ErrNoBudgetSet,
		},
		{
			name:      "DynamoDB error (BudgetSettings)",
			bsErr:     errors.New("dynamodb error"),
			wallet:    fixedW,
			wantErr:   errors.New("metrics: parallel read"),
		},
		{
			name:     "CountThisMonth error → 500",
			bs:       fixedBS,
			wallet:   fixedW,
			countErr: errors.New("dynamodb error"),
			wantErr:  errors.New("metrics: count_this_month"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := NewService(
				&mockBudgetSettingsReader{record: tc.bs, err: tc.bsErr},
				&mockWalletReader{record: tc.wallet, err: tc.walletErr},
				&mockOrderHistoryReader{count: tc.count, err: tc.countErr},
			)
			got, err := svc.GetMetrics(context.Background(), "u1")
			if tc.wantErr != nil {
				if err == nil {
					t.Fatalf("want error containing %q, got nil", tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.check != nil {
				tc.check(t, got)
			}
		})
	}
}

// --- PBT: P-E-PBT-01 ConsumptionRate 不変条件 ---

func TestComputeConsumptionRatePBT(t *testing.T) {
	properties := gopter.NewProperties(nil)
	properties.Property("ConsumptionRate は常に [0.0, 1.0] に収まる", prop.ForAll(
		func(budget, remaining uint32) bool {
			b := int(budget%100_000) + 1
			r := int(remaining) % (b + 1)
			rate := computeConsumptionRate(b, r)
			return rate >= 0.0 && rate <= 1.0
		},
		gen.UInt32(),
		gen.UInt32(),
	))
	properties.TestingRun(t)
}
