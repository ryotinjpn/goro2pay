package wallet

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/ryotinjpn/goro2pay/apps/api/internal/repo/budget_reset_log"
	"github.com/ryotinjpn/goro2pay/apps/api/internal/repo/budget_settings"
	"github.com/ryotinjpn/goro2pay/apps/api/internal/repo/idempotency"
	"github.com/ryotinjpn/goro2pay/apps/api/internal/repo/wallet_repo"
)

// WalletRepository は Service が依存する Wallet 操作の minimum interface。
//
// テスト容易性のため interface 抽出。Production 実装は
// `*wallet_repo.Repository` がそのまま満たす。
type WalletRepository interface {
	Get(ctx context.Context, userID string) (*wallet_repo.WalletRecord, error)
	Create(ctx context.Context, userID string, balance int) error
	DeductConditional(ctx context.Context, userID string, amount int) (newBalance int, err error)
	UpdateBalance(ctx context.Context, userID string, delta int) error
	SetBalance(ctx context.Context, userID string, balance int) error
	ListAllUserIDs(ctx context.Context) ([]string, error)
}

// BudgetSettingsRepository は Service が依存する BudgetSettings 操作の minimum interface。
type BudgetSettingsRepository interface {
	Get(ctx context.Context, userID string) (*budget_settings.BudgetSettings, error)
	Set(ctx context.Context, userID string, monthlyBudget int, effectiveFrom time.Time) error
}

// IdempotencyRepository は Service が依存する冪等性操作の minimum interface。
type IdempotencyRepository interface {
	TryAcquire(ctx context.Context, key string, payload []byte, ttl time.Duration) (acquired bool, existing *idempotency.IdempotencyRecord, err error)
	SaveResponse(ctx context.Context, key string, response []byte) error
}

// BudgetResetLogRepository は Service が依存する月次リセットログ操作の minimum interface。
type BudgetResetLogRepository interface {
	Get(ctx context.Context, resetDate, userID string) (*budget_reset_log.Log, error)
	Insert(ctx context.Context, log *budget_reset_log.Log) error
}

// Service は WalletService 本実装 (LC-BUDGET-02)。
type Service struct {
	walletRepo  WalletRepository
	settingsRep BudgetSettingsRepository
	idemRepo    IdempotencyRepository
	resetLogRep BudgetResetLogRepository
	now         func() time.Time
}

// NewService は API Lambda 用 Service を構築する (4 つの Repository すべて必須)。
func NewService(
	walletRepo WalletRepository,
	settingsRep BudgetSettingsRepository,
	idemRepo IdempotencyRepository,
	resetLogRep BudgetResetLogRepository,
) *Service {
	return &Service{
		walletRepo:  walletRepo,
		settingsRep: settingsRep,
		idemRepo:    idemRepo,
		resetLogRep: resetLogRep,
		now:         time.Now,
	}
}

// NewSchedulerService は Scheduler Lambda 用 Service を構築する (ResetAll のみ
// 利用するため idemRepo を要求しない、Code Review Important 5)。
//
// 本 Service の GetBalance / SetBudget / Deduct を呼ぶと panic するため、
// Scheduler Lambda 以外から使ってはならない。
func NewSchedulerService(
	walletRepo WalletRepository,
	settingsRep BudgetSettingsRepository,
	resetLogRep BudgetResetLogRepository,
) *Service {
	return &Service{
		walletRepo:  walletRepo,
		settingsRep: settingsRep,
		idemRepo:    nil, // ResetAll は idempotency を使わない
		resetLogRep: resetLogRep,
		now:         time.Now,
	}
}

// 確認: WalletService interface (凍結 IF §3.1) を実装する。
var _ WalletService = (*Service)(nil)

// SetClock はテスト用に時刻関数を上書きする。
func (s *Service) SetClock(now func() time.Time) {
	s.now = now
}

// GetBalance は UC-B-03 残高取得。Wallet + BudgetSettings を合成して WalletSnapshot を返す。
func (s *Service) GetBalance(ctx context.Context, userID string) (*WalletSnapshot, error) {
	if err := ValidateUserID(userID); err != nil {
		return nil, err
	}
	wallet, err := s.walletRepo.Get(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("wallet: get_balance: %w", err)
	}
	if wallet == nil {
		// Wallet 未作成 (予算未設定) はそのまま nil を返す。Handler 側で 404 等にマップする。
		return nil, nil
	}
	settings, err := s.settingsRep.Get(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("wallet: get_balance settings: %w", err)
	}
	monthlyBudget := 0
	if settings != nil {
		monthlyBudget = settings.MonthlyBudget
	}
	return &WalletSnapshot{
		UserID:        wallet.UserID,
		Balance:       wallet.Balance,
		MonthlyBudget: monthlyBudget,
		UpdatedAt:     wallet.UpdatedAt,
	}, nil
}

// SetBudget は UC-B-01 (初回) / UC-B-02 (変更) を統合的に処理する。
//
// 初回判定 (DR-B-01): **Wallet が NotFound** なら初回パス、それ以外は変更パス。
//   - BudgetSettings の有無は判定に使わない (FD §4.2)。万が一 BudgetSettings が
//     既存で Wallet だけ消えている運用ミス状態 (例: 月初リセット中に Wallet テーブル
//     だけ部分障害) では、初回パスで settings を上書き + Wallet 再作成して復旧する。
//     これは Code Review Important 6 で議論された corner case で、仕様レベルで
//     「Wallet 不在 = 初回」という単純化を採用している。
//
// 変更時の差分調整 (DR-B-02): delta > 0 → balance += delta、delta < 0 → min(balance, newBudget)。
// 部分失敗時はそのまま返す (P-REL-02、補償なし)。
func (s *Service) SetBudget(ctx context.Context, userID string, monthlyBudget int) error {
	if err := ValidateUserID(userID); err != nil {
		return err
	}
	if err := ValidateMonthlyBudget(monthlyBudget); err != nil {
		return err
	}
	now := s.now().UTC()

	wallet, err := s.walletRepo.Get(ctx, userID)
	if err != nil {
		return fmt.Errorf("wallet: setbudget get: %w", err)
	}
	if wallet == nil {
		// 初回パス (DR-B-01)
		if err := s.settingsRep.Set(ctx, userID, monthlyBudget, now); err != nil {
			return fmt.Errorf("wallet: setbudget settings: %w", err)
		}
		if err := s.walletRepo.Create(ctx, userID, monthlyBudget); err != nil {
			return fmt.Errorf("wallet: setbudget create: %w", err)
		}
		return nil
	}
	// 変更パス
	settings, err := s.settingsRep.Get(ctx, userID)
	if err != nil {
		return fmt.Errorf("wallet: setbudget get_settings: %w", err)
	}
	oldBudget := 0
	if settings != nil {
		oldBudget = settings.MonthlyBudget
	}
	delta := monthlyBudget - oldBudget

	if err := s.settingsRep.Set(ctx, userID, monthlyBudget, now); err != nil {
		return fmt.Errorf("wallet: setbudget settings_update: %w", err)
	}

	if delta > 0 {
		// 増額: 当月残高に delta を加算
		if err := s.walletRepo.UpdateBalance(ctx, userID, delta); err != nil {
			return fmt.Errorf("wallet: setbudget increase: %w", err)
		}
	} else if delta < 0 {
		// 減額: max(現残高, 新予算) ではなく min(現残高, 新予算) で打ち切り
		// (FD §4.2 注: 現残高が新予算より大きい場合のみ新予算まで切り下げる)
		newBalance := wallet.Balance
		if newBalance > monthlyBudget {
			newBalance = monthlyBudget
		}
		if err := s.walletRepo.SetBalance(ctx, userID, newBalance); err != nil {
			return fmt.Errorf("wallet: setbudget decrease: %w", err)
		}
	}
	// delta == 0 は no-op
	return nil
}

// deductPayload は冪等性 hash 計算用の固定 schema。
type deductPayload struct {
	Amount int `json:"amount"`
}

// deductResponse は IdempotencyRecord に保存する成功 / 失敗レスポンス。
type deductResponse struct {
	NewBalance int    `json:"newBalance,omitempty"`
	Error      string `json:"error,omitempty"`
}

// Deduct は UC-B-04 残高減算。冪等性キー単位で 1 回のみ残高を減らす。
//
// 処理フロー (FD §4.3):
//  1. validate (VR-B-03/04/05)
//  2. payload serialize → IdempotencyRepository.TryAcquire
//  3a. acquired=false かつ payload hash 一致 → 保存 response を再生 (DR-B-03)
//  3b. acquired=false かつ payload hash 不一致 → ErrIdempotencyConflict
//  3c. acquired=true → DeductConditional 実行
//      - 成功: SaveResponse(成功) → DeductResult 返却
//      - 残高不足: SaveResponse(失敗) → ErrInsufficientBalance (PR-B-04)
func (s *Service) Deduct(ctx context.Context, userID string, amount int, idempotencyKey string) (*DeductResult, error) {
	if err := ValidateDeductInput(userID, amount, idempotencyKey); err != nil {
		return nil, err
	}
	// json.Marshal は deductPayload (int のみ) では失敗しないため err 握りつぶし許容
	// (Code Review Minor 14: ポリシー統一)。
	payload, _ := json.Marshal(deductPayload{Amount: amount})

	acquired, existing, err := s.idemRepo.TryAcquire(ctx, idempotencyKey, payload, idempotency.DefaultTTL)
	if err != nil {
		return nil, fmt.Errorf("wallet: tryacquire: %w", err)
	}

	if !acquired {
		// 既存レコードあり: payload hash 一致確認 (DR-B-03)
		if !payloadEqual(existing.Payload, payload) {
			return nil, ErrIdempotencyConflict
		}
		// Response が空 (= 初回 PutItem 後の SaveResponse 失敗、または初回処理が
		// in-flight) の場合は in-progress として 503 にマップする。NewBalance:0 を
		// 「成功」として返してしまうと残高表示が破綻するため (Code Review Critical 2)、
		// クライアントは少し待って再試行する。
		if len(existing.Response) == 0 {
			return nil, ErrIdempotencyInProgress
		}
		// 保存済み response を再生
		var resp deductResponse
		if uerr := json.Unmarshal(existing.Response, &resp); uerr != nil {
			return nil, fmt.Errorf("wallet: unmarshal existing response: %w", uerr)
		}
		if resp.Error != "" {
			// 失敗結果を再生 (PR-B-04)
			if resp.Error == "ErrInsufficientBalance" {
				return nil, ErrInsufficientBalance
			}
			return nil, fmt.Errorf("wallet: replay error: %s", resp.Error)
		}
		// 成功 response 再生
		return &DeductResult{NewBalance: resp.NewBalance, Idempotent: true}, nil
	}

	// 新規取得: DeductConditional 実行
	newBalance, derr := s.walletRepo.DeductConditional(ctx, userID, amount)
	if derr != nil {
		// 失敗結果を IdempotencyRecord に保存して race condition で同じエラーを返せるようにする (PR-B-04)
		if errors.Is(derr, wallet_repo.ErrInsufficientBalance) {
			// json.Marshal は deductResponse (string + int) では失敗しないため err 握りつぶしを許容
			// (Code Review Minor 14: ポリシー統一)。
			failureResp, _ := json.Marshal(deductResponse{Error: "ErrInsufficientBalance"})
			_ = s.idemRepo.SaveResponse(ctx, idempotencyKey, failureResp)
			return nil, ErrInsufficientBalance
		}
		return nil, fmt.Errorf("wallet: deduct: %w", derr)
	}

	// json.Marshal は deductResponse (int のみ) では失敗しないため err 握りつぶしを許容 (Minor 14)。
	successResp, _ := json.Marshal(deductResponse{NewBalance: newBalance})
	if err := s.idemRepo.SaveResponse(ctx, idempotencyKey, successResp); err != nil {
		// SaveResponse 失敗は致命的ではない: 既に DynamoDB の残高は減算済み。
		// 同一キー再送時に既存レコードの response が空のため、再度 DeductConditional を
		// 試みて二重減算が起きる懸念があるが、ConditionExpression で wallet 不変条件は
		// 保たれる。ログ出力のみ行う。
		slog.WarnContext(ctx, "save response failed",
			"action", "deduct",
			"key", idempotencyKey,
			"error", err.Error(),
		)
	}
	return &DeductResult{NewBalance: newBalance, Idempotent: false}, nil
}

func payloadEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	ha := sha256.Sum256(a)
	hb := sha256.Sum256(b)
	return ha == hb
}

// ResetAll は UC-B-05 月初リセット。全ユーザを処理し、失敗を Errors に集約する (P-OBS-02)。
//
// 処理フロー:
//  1. resetDate = formatYYYYMM(now JST)
//  2. ListAllUserIDs
//  3. 各ユーザで:
//     - BudgetResetLog.Get で skip 判定 (DR-B-05)
//     - BudgetSettings.Get で settings 取得 (nil なら skip)
//     - Wallet.Get で prevBalance 取得
//     - Wallet.SetBalance(monthlyBudget) で完全リセット (PR-B-01)
//     - BudgetResetLog.Insert で履歴 + 冪等性確保 (CR-B-05)
//     - 失敗時は ERROR ログ + errors 配列に追加して継続
//  4. 集計 INFO ログ + ResetResult 返却
func (s *Service) ResetAll(ctx context.Context) (*ResetResult, error) {
	jst := time.FixedZone("Asia/Tokyo", 9*60*60)
	// resetDate は「リセットによって始まる新しい月」を JST で `YYYY-MM` 表記。
	// EventBridge cron が UTC 月末日 15:00 UTC (= JST 翌日 0:00 = JST 新月の 1 日 0:00)
	// に発火するため、`s.now()` は新月の 1 日付近を指す。BudgetResetLog の PK
	// `resetDate` はこの「新月」を識別子として使う (Code Review Minor 13)。
	resetDate := s.now().In(jst).Format("2006-01")

	userIDs, err := s.walletRepo.ListAllUserIDs(ctx)
	if err != nil {
		return nil, fmt.Errorf("wallet: list_all: %w", err)
	}

	var errs []error
	processed := 0
	for _, uid := range userIDs {
		if rerr := s.resetOneUser(ctx, resetDate, uid); rerr != nil {
			slog.ErrorContext(ctx, "reset failed for user",
				"action", "monthly_reset",
				"targetUserId", uid,
				"error", rerr.Error(),
			)
			errs = append(errs, fmt.Errorf("user %s: %w", uid, rerr))
			continue
		}
		processed++
	}

	slog.InfoContext(ctx, "monthly reset complete",
		"action", "monthly_reset",
		"resetDate", resetDate,
		"processedUsers", processed,
		"failedUsers", len(errs),
	)

	return &ResetResult{ProcessedUsers: processed, Errors: errs}, nil
}

func (s *Service) resetOneUser(ctx context.Context, resetDate, userID string) error {
	// DR-B-05 既処理スキップ
	existing, err := s.resetLogRep.Get(ctx, resetDate, userID)
	if err != nil {
		return fmt.Errorf("get_reset_log: %w", err)
	}
	if existing != nil {
		return nil // 既処理: 静かにスキップ
	}

	settings, err := s.settingsRep.Get(ctx, userID)
	if err != nil {
		return fmt.Errorf("get_settings: %w", err)
	}
	if settings == nil {
		return nil // 予算未設定: スキップ
	}

	wallet, err := s.walletRepo.Get(ctx, userID)
	if err != nil {
		return fmt.Errorf("get_wallet: %w", err)
	}
	if wallet == nil {
		return nil // Wallet 未作成: スキップ
	}

	prevBalance := wallet.Balance
	newBalance := settings.MonthlyBudget

	if err := s.walletRepo.SetBalance(ctx, userID, newBalance); err != nil {
		return fmt.Errorf("set_balance: %w", err)
	}

	if err := s.resetLogRep.Insert(ctx, &budget_reset_log.Log{
		ResetDate:   resetDate,
		UserID:      userID,
		PrevBalance: prevBalance,
		NewBalance:  newBalance,
		At:          s.now().UTC(),
	}); err != nil && !errors.Is(err, budget_reset_log.ErrAlreadyLogged) {
		return fmt.Errorf("insert_log: %w", err)
	}

	slog.InfoContext(ctx, "user reset",
		"action", "monthly_reset",
		"targetUserId", userID,
		"prevBalance", prevBalance,
		"newBalance", newBalance,
	)
	return nil
}
