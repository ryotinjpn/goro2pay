// Package budget_reset_log は BudgetResetLog テーブルへの DynamoDB アクセスを提供する。
//
// 月初リセット (UC-B-05) の冪等性を `(resetDate, userID)` 複合キー +
// `attribute_not_exists` ConditionExpression で保証する (CR-B-05 / P-REL-03)。
package budget_reset_log

import "time"

// Log は BudgetResetLog テーブルの 1 行に対応する。
type Log struct {
	ResetDate   string // YYYY-MM 形式
	UserID      string
	PrevBalance int
	NewBalance  int
	At          time.Time
}
