package budget_raise

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	"github.com/ryotinjpn/goro2pay/apps/api/internal/repo/budget_settings"
)

// --- mocks ---

type mockBSReader struct {
	record *budget_settings.BudgetSettings
	err    error
}

func (m *mockBSReader) Get(_ context.Context, _ string) (*budget_settings.BudgetSettings, error) {
	return m.record, m.err
}

type mockBSWriter struct {
	called bool
	err    error
}

func (m *mockBSWriter) Set(_ context.Context, _ string, _ int, _ time.Time) error {
	m.called = true
	return m.err
}

func fixedBS(budget int) *budget_settings.BudgetSettings {
	return &budget_settings.BudgetSettings{UserID: "u1", MonthlyBudget: budget, EffectiveFrom: time.Now()}
}

// --- ComputeRecommendedBudget ---

func TestComputeRecommendedBudget(t *testing.T) {
	cases := []struct {
		name    string
		current int
		want    int
		bsErr   error
		wantErr error
	}{
		{"通常値 30,000 → 45,000", 30_000, 45_000, nil, nil},
		{"上限クランプ 70,000 → 100,000", 70_000, 100_000, nil, nil},
		{"上限ちょうど 100,000 → 100,000", 100_000, 100_000, nil, nil},
		{"未設定 → ErrNoBudgetSet", 0, 0, nil, ErrNoBudgetSet},
		{"DynamoDB error", 0, 0, errors.New("db error"), errors.New("budget_raise")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var rec *budget_settings.BudgetSettings
			if tc.current > 0 {
				rec = fixedBS(tc.current)
			}
			svc := NewService(&mockBSReader{record: rec, err: tc.bsErr}, &mockBSWriter{})
			got, err := svc.ComputeRecommendedBudget(context.Background(), "u1")
			if tc.wantErr != nil {
				if err == nil {
					t.Fatalf("want error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("want %d got %d", tc.want, got)
			}
		})
	}
}

// --- Accept ---

func TestAccept(t *testing.T) {
	cases := []struct {
		name      string
		newBudget int
		bsRecord  *budget_settings.BudgetSettings
		writeErr  error
		wantErr   error
	}{
		{"正常", 45_000, fixedBS(30_000), nil, nil},
		{"範囲外 0 → ErrInvalidBudget", 0, fixedBS(30_000), nil, ErrInvalidBudget},
		{"範囲外 100_001 → ErrInvalidBudget", 100_001, fixedBS(30_000), nil, ErrInvalidBudget},
		{"未設定 → ErrNoBudgetSet", 45_000, nil, nil, ErrNoBudgetSet},
		{"Write error → 500", 45_000, fixedBS(30_000), errors.New("db"), errors.New("budget_raise: set")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := &mockBSWriter{err: tc.writeErr}
			svc := NewService(&mockBSReader{record: tc.bsRecord}, w)
			res, err := svc.Accept(context.Background(), "u1", tc.newBudget)
			if tc.wantErr != nil {
				if err == nil {
					t.Fatalf("want error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if res.NewMonthlyBudget != tc.newBudget {
				t.Errorf("want %d got %d", tc.newBudget, res.NewMonthlyBudget)
			}
			if !w.called {
				t.Error("BudgetSettingsWriter.Set not called")
			}
		})
	}
}

// --- PBT: P-E-PBT-02 ComputeRecommendedBudget 不変条件 ---

func TestComputeRecommendedBudgetPBT(t *testing.T) {
	properties := gopter.NewProperties(nil)
	properties.Property("推奨予算は現予算以上かつ 100,000 以下", prop.ForAll(
		func(current uint32) bool {
			c := int(current%100_000) + 1
			rec := computeRecommendedBudget(c)
			return rec >= c && rec <= maxBudget
		},
		gen.UInt32(),
	))
	properties.TestingRun(t)
}
