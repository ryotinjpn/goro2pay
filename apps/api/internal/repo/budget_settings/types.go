// Package budget_settings は BudgetSettings テーブルへの DynamoDB アクセスを提供する。
//
// 公開 IF (凍結 IF §3.2 / §6.3):
//   - BudgetSettingsReader: Unit E (MetricsService / BudgetRaiseService) が読取参照する subset
//   - BudgetSettingsWriter: Unit E (BudgetRaiseService.Accept) が翌月適用の予算更新で呼ぶ Writer
//   - BudgetSettings: 公開 DTO
//   - RaiseLog: 増額履歴エントリ
package budget_settings

import (
	"context"
	"time"
)

// BudgetSettings は BudgetSettings テーブルの 1 行に対応する公開 DTO (凍結 IF §3.2)。
type BudgetSettings struct {
	UserID        string
	MonthlyBudget int
	EffectiveFrom time.Time
}

// RaiseLog は raiseHistory に追記される増額履歴エントリ。
type RaiseLog struct {
	At         time.Time
	PrevBudget int
	NewBudget  int
}

// BudgetSettingsReader は Unit E が読取参照する subset (凍結 IF §3.2)。
type BudgetSettingsReader interface {
	Get(ctx context.Context, userID string) (*BudgetSettings, error)
}

// BudgetSettingsWriter は Unit E (BudgetRaiseService.Accept) が呼ぶ Writer (凍結 IF §6.3)。
//
// 翌月 1 日 00:00 JST を effectiveFrom として渡される (Unit B 内部の SetBudget は
// 即時反映なのでこの Writer は使わない)。
type BudgetSettingsWriter interface {
	Set(ctx context.Context, userID string, monthlyBudget int, effectiveFrom time.Time) error
}
