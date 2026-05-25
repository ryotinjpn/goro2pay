package wallet

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/ryotinjpn/goro2pay/apps/api/internal/repo/budget_reset_log"
	"github.com/ryotinjpn/goro2pay/apps/api/internal/repo/budget_settings"
	"github.com/ryotinjpn/goro2pay/apps/api/internal/repo/idempotency"
	"github.com/ryotinjpn/goro2pay/apps/api/internal/repo/wallet_repo"
)

// ----- function-field mocks -----

type mockWalletRepo struct {
	getFn               func(ctx context.Context, userID string) (*wallet_repo.WalletRecord, error)
	createFn            func(ctx context.Context, userID string, balance int) error
	deductConditionalFn func(ctx context.Context, userID string, amount int) (int, error)
	updateBalanceFn     func(ctx context.Context, userID string, delta int) error
	setBalanceFn       func(ctx context.Context, userID string, balance int) error
	listAllUserIDsFn   func(ctx context.Context) ([]string, error)
}

func (m *mockWalletRepo) Get(ctx context.Context, userID string) (*wallet_repo.WalletRecord, error) {
	return m.getFn(ctx, userID)
}
func (m *mockWalletRepo) Create(ctx context.Context, userID string, balance int) error {
	return m.createFn(ctx, userID, balance)
}
func (m *mockWalletRepo) DeductConditional(ctx context.Context, userID string, amount int) (int, error) {
	return m.deductConditionalFn(ctx, userID, amount)
}
func (m *mockWalletRepo) UpdateBalance(ctx context.Context, userID string, delta int) error {
	return m.updateBalanceFn(ctx, userID, delta)
}
func (m *mockWalletRepo) SetBalance(ctx context.Context, userID string, balance int) error {
	return m.setBalanceFn(ctx, userID, balance)
}
func (m *mockWalletRepo) ListAllUserIDs(ctx context.Context) ([]string, error) {
	return m.listAllUserIDsFn(ctx)
}

type mockSettingsRepo struct {
	getFn func(ctx context.Context, userID string) (*budget_settings.BudgetSettings, error)
	setFn func(ctx context.Context, userID string, monthlyBudget int, effectiveFrom time.Time) error
}

func (m *mockSettingsRepo) Get(ctx context.Context, userID string) (*budget_settings.BudgetSettings, error) {
	return m.getFn(ctx, userID)
}
func (m *mockSettingsRepo) Set(ctx context.Context, userID string, monthlyBudget int, effectiveFrom time.Time) error {
	return m.setFn(ctx, userID, monthlyBudget, effectiveFrom)
}

type mockIdemRepo struct {
	tryAcquireFn   func(ctx context.Context, key string, payload []byte, ttl time.Duration) (bool, *idempotency.IdempotencyRecord, error)
	saveResponseFn func(ctx context.Context, key string, response []byte) error
}

func (m *mockIdemRepo) TryAcquire(ctx context.Context, key string, payload []byte, ttl time.Duration) (bool, *idempotency.IdempotencyRecord, error) {
	return m.tryAcquireFn(ctx, key, payload, ttl)
}
func (m *mockIdemRepo) SaveResponse(ctx context.Context, key string, response []byte) error {
	if m.saveResponseFn == nil {
		return nil
	}
	return m.saveResponseFn(ctx, key, response)
}

type mockResetLogRepo struct {
	getFn    func(ctx context.Context, resetDate, userID string) (*budget_reset_log.Log, error)
	insertFn func(ctx context.Context, log *budget_reset_log.Log) error
}

func (m *mockResetLogRepo) Get(ctx context.Context, resetDate, userID string) (*budget_reset_log.Log, error) {
	return m.getFn(ctx, resetDate, userID)
}
func (m *mockResetLogRepo) Insert(ctx context.Context, log *budget_reset_log.Log) error {
	return m.insertFn(ctx, log)
}

func newSvc(w WalletRepository, s BudgetSettingsRepository, i IdempotencyRepository, r BudgetResetLogRepository) *Service {
	svc := NewService(w, s, i, r)
	svc.SetClock(func() time.Time { return time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC) })
	return svc
}

// ----- GetBalance -----

func TestService_GetBalance_OK(t *testing.T) {
	w := &mockWalletRepo{
		getFn: func(ctx context.Context, userID string) (*wallet_repo.WalletRecord, error) {
			return &wallet_repo.WalletRecord{UserID: userID, Balance: 28800, UpdatedAt: time.Now().UTC()}, nil
		},
	}
	s := &mockSettingsRepo{
		getFn: func(ctx context.Context, userID string) (*budget_settings.BudgetSettings, error) {
			return &budget_settings.BudgetSettings{UserID: userID, MonthlyBudget: 30000}, nil
		},
	}
	svc := newSvc(w, s, &mockIdemRepo{}, &mockResetLogRepo{})
	got, err := svc.GetBalance(context.Background(), "user_a")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got == nil {
		t.Fatal("expected snapshot, got nil")
	}
	if got.Balance != 28800 || got.MonthlyBudget != 30000 {
		t.Fatalf("unexpected snapshot: %+v", got)
	}
}

func TestService_GetBalance_WalletNotFound(t *testing.T) {
	w := &mockWalletRepo{
		getFn: func(ctx context.Context, userID string) (*wallet_repo.WalletRecord, error) {
			return nil, nil
		},
	}
	svc := newSvc(w, &mockSettingsRepo{}, &mockIdemRepo{}, &mockResetLogRepo{})
	got, err := svc.GetBalance(context.Background(), "user_a")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil snapshot, got %+v", got)
	}
}

// ----- SetBudget -----

func TestService_SetBudget_FirstTime(t *testing.T) {
	createCalled := false
	settingsSetCalled := false
	w := &mockWalletRepo{
		getFn: func(ctx context.Context, userID string) (*wallet_repo.WalletRecord, error) {
			return nil, nil // 初回判定
		},
		createFn: func(ctx context.Context, userID string, balance int) error {
			createCalled = true
			if balance != 30000 {
				t.Fatalf("expected balance=30000, got %d", balance)
			}
			return nil
		},
	}
	s := &mockSettingsRepo{
		setFn: func(ctx context.Context, userID string, monthlyBudget int, effectiveFrom time.Time) error {
			settingsSetCalled = true
			if monthlyBudget != 30000 {
				t.Fatalf("expected monthlyBudget=30000, got %d", monthlyBudget)
			}
			return nil
		},
	}
	svc := newSvc(w, s, &mockIdemRepo{}, &mockResetLogRepo{})
	if err := svc.SetBudget(context.Background(), "user_a", 30000); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !createCalled || !settingsSetCalled {
		t.Fatal("expected both Create and Settings.Set to be called")
	}
}

func TestService_SetBudget_Increase(t *testing.T) {
	updateBalanceDelta := -1
	w := &mockWalletRepo{
		getFn: func(ctx context.Context, userID string) (*wallet_repo.WalletRecord, error) {
			return &wallet_repo.WalletRecord{UserID: userID, Balance: 18000}, nil
		},
		updateBalanceFn: func(ctx context.Context, userID string, delta int) error {
			updateBalanceDelta = delta
			return nil
		},
		setBalanceFn: func(ctx context.Context, userID string, balance int) error {
			t.Fatalf("SetBalance should not be called for increase, got balance=%d", balance)
			return nil
		},
	}
	s := &mockSettingsRepo{
		getFn: func(ctx context.Context, userID string) (*budget_settings.BudgetSettings, error) {
			return &budget_settings.BudgetSettings{UserID: userID, MonthlyBudget: 30000}, nil
		},
		setFn: func(ctx context.Context, userID string, monthlyBudget int, effectiveFrom time.Time) error {
			return nil
		},
	}
	svc := newSvc(w, s, &mockIdemRepo{}, &mockResetLogRepo{})
	if err := svc.SetBudget(context.Background(), "user_a", 50000); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if updateBalanceDelta != 20000 {
		t.Fatalf("expected delta=+20000, got %d", updateBalanceDelta)
	}
}

func TestService_SetBudget_Decrease_Truncate(t *testing.T) {
	// balance(25000) > newBudget(20000) → SetBalance(20000) で打ち切り
	setBalanceCalled := -1
	w := &mockWalletRepo{
		getFn: func(ctx context.Context, userID string) (*wallet_repo.WalletRecord, error) {
			return &wallet_repo.WalletRecord{UserID: userID, Balance: 25000}, nil
		},
		setBalanceFn: func(ctx context.Context, userID string, balance int) error {
			setBalanceCalled = balance
			return nil
		},
		updateBalanceFn: func(ctx context.Context, userID string, delta int) error {
			t.Fatalf("UpdateBalance should not be called for decrease, got delta=%d", delta)
			return nil
		},
	}
	s := &mockSettingsRepo{
		getFn: func(ctx context.Context, userID string) (*budget_settings.BudgetSettings, error) {
			return &budget_settings.BudgetSettings{UserID: userID, MonthlyBudget: 30000}, nil
		},
		setFn: func(ctx context.Context, userID string, monthlyBudget int, effectiveFrom time.Time) error {
			return nil
		},
	}
	svc := newSvc(w, s, &mockIdemRepo{}, &mockResetLogRepo{})
	if err := svc.SetBudget(context.Background(), "user_a", 20000); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if setBalanceCalled != 20000 {
		t.Fatalf("expected SetBalance(20000), got %d", setBalanceCalled)
	}
}

func TestService_SetBudget_Decrease_NoTruncate(t *testing.T) {
	// balance(15000) <= newBudget(20000) → SetBalance(15000) で維持
	setBalanceCalled := -1
	w := &mockWalletRepo{
		getFn: func(ctx context.Context, userID string) (*wallet_repo.WalletRecord, error) {
			return &wallet_repo.WalletRecord{UserID: userID, Balance: 15000}, nil
		},
		setBalanceFn: func(ctx context.Context, userID string, balance int) error {
			setBalanceCalled = balance
			return nil
		},
	}
	s := &mockSettingsRepo{
		getFn: func(ctx context.Context, userID string) (*budget_settings.BudgetSettings, error) {
			return &budget_settings.BudgetSettings{UserID: userID, MonthlyBudget: 30000}, nil
		},
		setFn: func(ctx context.Context, userID string, monthlyBudget int, effectiveFrom time.Time) error {
			return nil
		},
	}
	svc := newSvc(w, s, &mockIdemRepo{}, &mockResetLogRepo{})
	if err := svc.SetBudget(context.Background(), "user_a", 20000); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if setBalanceCalled != 15000 {
		t.Fatalf("expected SetBalance(15000) for no truncation, got %d", setBalanceCalled)
	}
}

func TestService_SetBudget_OutOfRange(t *testing.T) {
	svc := newSvc(&mockWalletRepo{}, &mockSettingsRepo{}, &mockIdemRepo{}, &mockResetLogRepo{})
	err := svc.SetBudget(context.Background(), "user_a", 500)
	if !errors.Is(err, ErrBudgetOutOfRange) {
		t.Fatalf("expected ErrBudgetOutOfRange, got %v", err)
	}
}

func TestService_SetBudget_NotMultipleOf1000(t *testing.T) {
	svc := newSvc(&mockWalletRepo{}, &mockSettingsRepo{}, &mockIdemRepo{}, &mockResetLogRepo{})
	err := svc.SetBudget(context.Background(), "user_a", 30001)
	if !errors.Is(err, ErrBudgetOutOfRange) {
		t.Fatalf("expected ErrBudgetOutOfRange, got %v", err)
	}
}

// ----- Deduct -----

const validKey = "user_a:01HW3XKQ7EZ8XN1MX1C7Q4Z9TY"

func TestService_Deduct_Success(t *testing.T) {
	w := &mockWalletRepo{
		deductConditionalFn: func(ctx context.Context, userID string, amount int) (int, error) {
			return 29150, nil
		},
	}
	saveResp := []byte(nil)
	idem := &mockIdemRepo{
		tryAcquireFn: func(ctx context.Context, key string, payload []byte, ttl time.Duration) (bool, *idempotency.IdempotencyRecord, error) {
			return true, nil, nil
		},
		saveResponseFn: func(ctx context.Context, key string, response []byte) error {
			saveResp = response
			return nil
		},
	}
	svc := newSvc(w, &mockSettingsRepo{}, idem, &mockResetLogRepo{})
	got, err := svc.Deduct(context.Background(), "user_a", 850, validKey)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got.NewBalance != 29150 || got.Idempotent {
		t.Fatalf("unexpected result: %+v", got)
	}
	if len(saveResp) == 0 {
		t.Fatal("expected SaveResponse to be called with non-empty payload")
	}
}

func TestService_Deduct_InsufficientBalance(t *testing.T) {
	failureSaved := false
	w := &mockWalletRepo{
		deductConditionalFn: func(ctx context.Context, userID string, amount int) (int, error) {
			return 0, wallet_repo.ErrInsufficientBalance
		},
	}
	idem := &mockIdemRepo{
		tryAcquireFn: func(ctx context.Context, key string, payload []byte, ttl time.Duration) (bool, *idempotency.IdempotencyRecord, error) {
			return true, nil, nil
		},
		saveResponseFn: func(ctx context.Context, key string, response []byte) error {
			var r struct {
				Error string `json:"error"`
			}
			_ = json.Unmarshal(response, &r)
			if r.Error == "ErrInsufficientBalance" {
				failureSaved = true
			}
			return nil
		},
	}
	svc := newSvc(w, &mockSettingsRepo{}, idem, &mockResetLogRepo{})
	_, err := svc.Deduct(context.Background(), "user_a", 5000, validKey)
	if !errors.Is(err, ErrInsufficientBalance) {
		t.Fatalf("expected ErrInsufficientBalance, got %v", err)
	}
	if !failureSaved {
		t.Fatal("expected failure response to be saved (PR-B-04)")
	}
}

func TestService_Deduct_IdempotencyHit_SamePayload(t *testing.T) {
	prevPayload, _ := json.Marshal(deductPayload{Amount: 850})
	prevResp, _ := json.Marshal(deductResponse{NewBalance: 29150})
	deductCalled := false
	w := &mockWalletRepo{
		deductConditionalFn: func(ctx context.Context, userID string, amount int) (int, error) {
			deductCalled = true
			return 0, nil
		},
	}
	idem := &mockIdemRepo{
		tryAcquireFn: func(ctx context.Context, key string, payload []byte, ttl time.Duration) (bool, *idempotency.IdempotencyRecord, error) {
			return false, &idempotency.IdempotencyRecord{Payload: prevPayload, Response: prevResp}, nil
		},
	}
	svc := newSvc(w, &mockSettingsRepo{}, idem, &mockResetLogRepo{})
	got, err := svc.Deduct(context.Background(), "user_a", 850, validKey)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !got.Idempotent || got.NewBalance != 29150 {
		t.Fatalf("expected idempotent replay, got %+v", got)
	}
	if deductCalled {
		t.Fatal("DeductConditional should not be called on idempotent hit")
	}
}

func TestService_Deduct_IdempotencyConflict_DifferentPayload(t *testing.T) {
	prevPayload, _ := json.Marshal(deductPayload{Amount: 500})
	idem := &mockIdemRepo{
		tryAcquireFn: func(ctx context.Context, key string, payload []byte, ttl time.Duration) (bool, *idempotency.IdempotencyRecord, error) {
			return false, &idempotency.IdempotencyRecord{Payload: prevPayload, Response: nil}, nil
		},
	}
	svc := newSvc(&mockWalletRepo{}, &mockSettingsRepo{}, idem, &mockResetLogRepo{})
	_, err := svc.Deduct(context.Background(), "user_a", 850, validKey)
	if !errors.Is(err, ErrIdempotencyConflict) {
		t.Fatalf("expected ErrIdempotencyConflict, got %v", err)
	}
}

func TestService_Deduct_FailureReplay(t *testing.T) {
	prevPayload, _ := json.Marshal(deductPayload{Amount: 5000})
	prevResp, _ := json.Marshal(deductResponse{Error: "ErrInsufficientBalance"})
	idem := &mockIdemRepo{
		tryAcquireFn: func(ctx context.Context, key string, payload []byte, ttl time.Duration) (bool, *idempotency.IdempotencyRecord, error) {
			return false, &idempotency.IdempotencyRecord{Payload: prevPayload, Response: prevResp}, nil
		},
	}
	svc := newSvc(&mockWalletRepo{}, &mockSettingsRepo{}, idem, &mockResetLogRepo{})
	_, err := svc.Deduct(context.Background(), "user_a", 5000, validKey)
	if !errors.Is(err, ErrInsufficientBalance) {
		t.Fatalf("expected ErrInsufficientBalance replay, got %v", err)
	}
}

// Critical 2 修正検証: Response 空 (in-flight) の状態で再送 → ErrIdempotencyInProgress
func TestService_Deduct_IdempotencyInProgress(t *testing.T) {
	prevPayload, _ := json.Marshal(deductPayload{Amount: 850})
	idem := &mockIdemRepo{
		tryAcquireFn: func(ctx context.Context, key string, payload []byte, ttl time.Duration) (bool, *idempotency.IdempotencyRecord, error) {
			// Response が nil/空 = 初回処理が in-flight、または SaveResponse 失敗状態
			return false, &idempotency.IdempotencyRecord{Payload: prevPayload, Response: nil}, nil
		},
	}
	w := &mockWalletRepo{
		deductConditionalFn: func(ctx context.Context, userID string, amount int) (int, error) {
			t.Fatal("DeductConditional should not be called when in-progress")
			return 0, nil
		},
	}
	svc := newSvc(w, &mockSettingsRepo{}, idem, &mockResetLogRepo{})
	_, err := svc.Deduct(context.Background(), "user_a", 850, validKey)
	if !errors.Is(err, ErrIdempotencyInProgress) {
		t.Fatalf("expected ErrIdempotencyInProgress, got %v", err)
	}
}

func TestService_Deduct_InvalidInput_BadKey(t *testing.T) {
	svc := newSvc(&mockWalletRepo{}, &mockSettingsRepo{}, &mockIdemRepo{}, &mockResetLogRepo{})
	_, err := svc.Deduct(context.Background(), "user_a", 100, "invalid_key")
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestService_Deduct_InvalidInput_KeyMismatch(t *testing.T) {
	svc := newSvc(&mockWalletRepo{}, &mockSettingsRepo{}, &mockIdemRepo{}, &mockResetLogRepo{})
	_, err := svc.Deduct(context.Background(), "user_b", 100, validKey)
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

// ----- ResetAll -----

func TestService_ResetAll_AllSuccess(t *testing.T) {
	users := []string{"user_a", "user_b"}
	setBalanceCalls := []int{}
	w := &mockWalletRepo{
		listAllUserIDsFn: func(ctx context.Context) ([]string, error) { return users, nil },
		getFn: func(ctx context.Context, userID string) (*wallet_repo.WalletRecord, error) {
			return &wallet_repo.WalletRecord{UserID: userID, Balance: 5000}, nil
		},
		setBalanceFn: func(ctx context.Context, userID string, balance int) error {
			setBalanceCalls = append(setBalanceCalls, balance)
			return nil
		},
	}
	s := &mockSettingsRepo{
		getFn: func(ctx context.Context, userID string) (*budget_settings.BudgetSettings, error) {
			return &budget_settings.BudgetSettings{UserID: userID, MonthlyBudget: 30000}, nil
		},
	}
	r := &mockResetLogRepo{
		getFn:    func(ctx context.Context, resetDate, userID string) (*budget_reset_log.Log, error) { return nil, nil },
		insertFn: func(ctx context.Context, log *budget_reset_log.Log) error { return nil },
	}
	svc := newSvc(w, s, &mockIdemRepo{}, r)
	res, err := svc.ResetAll(context.Background())
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if res.ProcessedUsers != 2 || len(res.Errors) != 0 {
		t.Fatalf("unexpected result: %+v", res)
	}
	if len(setBalanceCalls) != 2 || setBalanceCalls[0] != 30000 {
		t.Fatalf("expected 2 SetBalance(30000) calls, got %v", setBalanceCalls)
	}
}

func TestService_ResetAll_PartialFailure(t *testing.T) {
	users := []string{"user_a", "user_b"}
	w := &mockWalletRepo{
		listAllUserIDsFn: func(ctx context.Context) ([]string, error) { return users, nil },
		getFn: func(ctx context.Context, userID string) (*wallet_repo.WalletRecord, error) {
			return &wallet_repo.WalletRecord{UserID: userID, Balance: 5000}, nil
		},
		setBalanceFn: func(ctx context.Context, userID string, balance int) error {
			if userID == "user_b" {
				return errors.New("simulated dynamodb failure")
			}
			return nil
		},
	}
	s := &mockSettingsRepo{
		getFn: func(ctx context.Context, userID string) (*budget_settings.BudgetSettings, error) {
			return &budget_settings.BudgetSettings{UserID: userID, MonthlyBudget: 30000}, nil
		},
	}
	r := &mockResetLogRepo{
		getFn:    func(ctx context.Context, resetDate, userID string) (*budget_reset_log.Log, error) { return nil, nil },
		insertFn: func(ctx context.Context, log *budget_reset_log.Log) error { return nil },
	}
	svc := newSvc(w, s, &mockIdemRepo{}, r)
	res, err := svc.ResetAll(context.Background())
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if res.ProcessedUsers != 1 || len(res.Errors) != 1 {
		t.Fatalf("expected 1 success + 1 error, got %+v", res)
	}
}

func TestService_ResetAll_SkipExisting(t *testing.T) {
	users := []string{"user_a"}
	setBalanceCalled := false
	w := &mockWalletRepo{
		listAllUserIDsFn: func(ctx context.Context) ([]string, error) { return users, nil },
		setBalanceFn: func(ctx context.Context, userID string, balance int) error {
			setBalanceCalled = true
			return nil
		},
	}
	r := &mockResetLogRepo{
		getFn: func(ctx context.Context, resetDate, userID string) (*budget_reset_log.Log, error) {
			return &budget_reset_log.Log{ResetDate: resetDate, UserID: userID}, nil
		},
	}
	svc := newSvc(w, &mockSettingsRepo{}, &mockIdemRepo{}, r)
	res, err := svc.ResetAll(context.Background())
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if res.ProcessedUsers != 1 || len(res.Errors) != 0 {
		t.Fatalf("expected processed=1 errors=0, got %+v", res)
	}
	if setBalanceCalled {
		t.Fatal("SetBalance should not be called when reset log already exists")
	}
}

func TestService_ResetAll_SkipNoSettings(t *testing.T) {
	users := []string{"user_a"}
	setBalanceCalled := false
	w := &mockWalletRepo{
		listAllUserIDsFn: func(ctx context.Context) ([]string, error) { return users, nil },
		setBalanceFn: func(ctx context.Context, userID string, balance int) error {
			setBalanceCalled = true
			return nil
		},
	}
	s := &mockSettingsRepo{
		getFn: func(ctx context.Context, userID string) (*budget_settings.BudgetSettings, error) {
			return nil, nil // BudgetSettings 未設定
		},
	}
	r := &mockResetLogRepo{
		getFn: func(ctx context.Context, resetDate, userID string) (*budget_reset_log.Log, error) { return nil, nil },
	}
	svc := newSvc(w, s, &mockIdemRepo{}, r)
	res, err := svc.ResetAll(context.Background())
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if res.ProcessedUsers != 1 || len(res.Errors) != 0 {
		t.Fatalf("expected processed=1 errors=0, got %+v", res)
	}
	if setBalanceCalled {
		t.Fatal("SetBalance should not be called when settings is nil")
	}
}
