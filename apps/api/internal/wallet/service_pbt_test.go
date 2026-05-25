package wallet

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	"github.com/ryotinjpn/goro2pay/apps/api/internal/repo/budget_reset_log"
	"github.com/ryotinjpn/goro2pay/apps/api/internal/repo/budget_settings"
	"github.com/ryotinjpn/goro2pay/apps/api/internal/repo/idempotency"
	"github.com/ryotinjpn/goro2pay/apps/api/internal/repo/wallet_repo"
)

// inmemoryWalletRepo は PBT 用の単純な in-memory Wallet repo (state 保持)。
type inmemoryWalletRepo struct {
	mu       sync.Mutex
	balances map[string]int
}

func newInmemoryWalletRepo() *inmemoryWalletRepo {
	return &inmemoryWalletRepo{balances: map[string]int{}}
}

func (r *inmemoryWalletRepo) Get(ctx context.Context, userID string) (*wallet_repo.WalletRecord, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	bal, ok := r.balances[userID]
	if !ok {
		return nil, nil
	}
	return &wallet_repo.WalletRecord{UserID: userID, Balance: bal, UpdatedAt: time.Now().UTC()}, nil
}
func (r *inmemoryWalletRepo) Create(ctx context.Context, userID string, balance int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.balances[userID] = balance
	return nil
}
func (r *inmemoryWalletRepo) DeductConditional(ctx context.Context, userID string, amount int) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	bal, ok := r.balances[userID]
	if !ok {
		return 0, wallet_repo.ErrInsufficientBalance
	}
	if bal < amount {
		return 0, wallet_repo.ErrInsufficientBalance
	}
	r.balances[userID] = bal - amount
	return r.balances[userID], nil
}
func (r *inmemoryWalletRepo) UpdateBalance(ctx context.Context, userID string, delta int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.balances[userID] += delta
	return nil
}
func (r *inmemoryWalletRepo) SetBalance(ctx context.Context, userID string, balance int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.balances[userID] = balance
	return nil
}
func (r *inmemoryWalletRepo) ListAllUserIDs(ctx context.Context) ([]string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	uids := make([]string, 0, len(r.balances))
	for k := range r.balances {
		uids = append(uids, k)
	}
	return uids, nil
}

// inmemorySettingsRepo は PBT 用の in-memory BudgetSettings repo。
type inmemorySettingsRepo struct {
	mu    sync.Mutex
	store map[string]*budget_settings.BudgetSettings
}

func newInmemorySettingsRepo() *inmemorySettingsRepo {
	return &inmemorySettingsRepo{store: map[string]*budget_settings.BudgetSettings{}}
}
func (r *inmemorySettingsRepo) Get(ctx context.Context, userID string) (*budget_settings.BudgetSettings, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	s, ok := r.store[userID]
	if !ok {
		return nil, nil
	}
	return s, nil
}
func (r *inmemorySettingsRepo) Set(ctx context.Context, userID string, monthlyBudget int, effectiveFrom time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.store[userID] = &budget_settings.BudgetSettings{
		UserID: userID, MonthlyBudget: monthlyBudget, EffectiveFrom: effectiveFrom,
	}
	return nil
}

// inmemoryIdemRepo は PBT 用の in-memory Idempotency repo。
type inmemoryIdemRepo struct {
	mu      sync.Mutex
	records map[string]*idempotency.IdempotencyRecord
}

func newInmemoryIdemRepo() *inmemoryIdemRepo {
	return &inmemoryIdemRepo{records: map[string]*idempotency.IdempotencyRecord{}}
}
func (r *inmemoryIdemRepo) TryAcquire(ctx context.Context, key string, payload []byte, ttl time.Duration) (bool, *idempotency.IdempotencyRecord, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if existing, ok := r.records[key]; ok {
		return false, existing, nil
	}
	r.records[key] = &idempotency.IdempotencyRecord{Key: key, Payload: payload, CreatedAt: time.Now()}
	return true, nil, nil
}
func (r *inmemoryIdemRepo) SaveResponse(ctx context.Context, key string, response []byte) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if rec, ok := r.records[key]; ok {
		rec.Response = response
	}
	return nil
}

// inmemoryResetLogRepo は PBT 用の in-memory BudgetResetLog repo。
type inmemoryResetLogRepo struct {
	mu    sync.Mutex
	store map[string]*budget_reset_log.Log
}

func newInmemoryResetLogRepo() *inmemoryResetLogRepo {
	return &inmemoryResetLogRepo{store: map[string]*budget_reset_log.Log{}}
}
func (r *inmemoryResetLogRepo) Get(ctx context.Context, resetDate, userID string) (*budget_reset_log.Log, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.store[resetDate+"#"+userID], nil
}
func (r *inmemoryResetLogRepo) Insert(ctx context.Context, log *budget_reset_log.Log) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.store[log.ResetDate+"#"+log.UserID] = log
	return nil
}

// makeSvcForPBT は in-memory な repos で組み立てた Service を返す。
func makeSvcForPBT() (*Service, *inmemoryWalletRepo, *inmemorySettingsRepo) {
	w := newInmemoryWalletRepo()
	s := newInmemorySettingsRepo()
	i := newInmemoryIdemRepo()
	r := newInmemoryResetLogRepo()
	svc := NewService(w, s, i, r)
	svc.SetClock(func() time.Time { return time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC) })
	return svc, w, s
}

// Property 1: 残高不変条件
//
// 任意の (initialBalance, amount) について:
//  - initialBalance >= amount のとき Deduct が成功し newBalance == initialBalance - amount
//  - initialBalance < amount のとき ErrInsufficientBalance を返し残高は変わらない
func TestProperty_BalanceInvariant(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("Deduct preserves balance invariant",
		prop.ForAll(
			func(initialBalance, amount int) bool {
				if amount <= 0 {
					return true // VR-B-03 で弾かれるので skip
				}
				svc, w, _ := makeSvcForPBT()
				ctx := context.Background()
				const userID = "user_a"
				const key = "user_a:01HW3XKQ7EZ8XN1MX1C7Q4Z9TY"

				// 初期残高をセット (Create)
				if err := w.Create(ctx, userID, initialBalance); err != nil {
					return false
				}

				res, err := svc.Deduct(ctx, userID, amount, key)
				if initialBalance >= amount {
					// 成功するはず
					if err != nil {
						return false
					}
					if res.NewBalance != initialBalance-amount {
						return false
					}
					// 残高が実際に減っている
					rec, _ := w.Get(ctx, userID)
					return rec.Balance == initialBalance-amount
				}
				// initialBalance < amount: ErrInsufficientBalance
				if !errors.Is(err, ErrInsufficientBalance) {
					return false
				}
				rec, _ := w.Get(ctx, userID)
				return rec.Balance == initialBalance
			},
			gen.IntRange(0, 1_000_000),
			gen.IntRange(1, 200_000),
		))

	properties.TestingRun(t)
}

// Property 2: SetBudget べき等性
//
// 任意の有効な monthlyBudget で SetBudget(x); SetBudget(x) の最終状態が、1 回呼び出しと同じ。
func TestProperty_SetBudgetIdempotency(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("SetBudget(x); SetBudget(x) == SetBudget(x)",
		prop.ForAll(
			func(monthlyBudgetMul int) bool {
				if monthlyBudgetMul < 1 || monthlyBudgetMul > 100 {
					return true
				}
				monthlyBudget := monthlyBudgetMul * 1000
				svc1, w1, _ := makeSvcForPBT()
				svc2, w2, _ := makeSvcForPBT()
				ctx := context.Background()
				const userID = "user_a"

				if err := svc1.SetBudget(ctx, userID, monthlyBudget); err != nil {
					return false
				}
				if err := svc2.SetBudget(ctx, userID, monthlyBudget); err != nil {
					return false
				}
				if err := svc2.SetBudget(ctx, userID, monthlyBudget); err != nil {
					return false
				}

				r1, _ := w1.Get(ctx, userID)
				r2, _ := w2.Get(ctx, userID)
				return r1 != nil && r2 != nil && r1.Balance == r2.Balance
			},
			gen.IntRange(1, 100),
		))

	properties.TestingRun(t)
}

// Property 3: バリデーション境界値
//
// 任意の int で:
//  - 1000 <= x <= 100000 かつ x % 1000 == 0 → nil
//  - それ以外 → ErrBudgetOutOfRange
func TestProperty_ValidateMonthlyBudgetBoundary(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("ValidateMonthlyBudget enforces range and step",
		prop.ForAll(
			func(x int) bool {
				err := ValidateMonthlyBudget(x)
				validRange := x >= MinMonthlyBudget && x <= MaxMonthlyBudget
				validStep := x%BudgetStep == 0
				if validRange && validStep {
					return err == nil
				}
				return errors.Is(err, ErrBudgetOutOfRange)
			},
			gen.IntRange(-200_000, 200_000),
		))

	properties.TestingRun(t)
}
