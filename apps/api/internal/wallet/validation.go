package wallet

import (
	"errors"
	"regexp"
	"strings"
)

// ErrInvalidInput は VR-B-03〜06 のバリデーション違反で返される (Handler 層で 400)。
var ErrInvalidInput = errors.New("wallet: invalid input")

const (
	// MinMonthlyBudget は VR-B-01 の下限 (円)。
	MinMonthlyBudget = 1_000
	// MaxMonthlyBudget は VR-B-01 の上限 (円)。
	MaxMonthlyBudget = 100_000
	// BudgetStep は VR-B-02 の刻み (円)。
	BudgetStep = 1_000
)

// idempotencyKeyPattern は VR-B-04 の `{userID}:{ulid}` 形式を検証する。
// ULID は 26 文字の Crockford's Base32。
var idempotencyKeyPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+:[0-9A-HJKMNP-TV-Z]{26}$`)

// ValidateMonthlyBudget は VR-B-01 (範囲) + VR-B-02 (1,000 円刻み) を検証する。
//
// 違反時は ErrBudgetOutOfRange を返す。
func ValidateMonthlyBudget(monthlyBudget int) error {
	if monthlyBudget < MinMonthlyBudget || monthlyBudget > MaxMonthlyBudget {
		return ErrBudgetOutOfRange
	}
	if monthlyBudget%BudgetStep != 0 {
		return ErrBudgetOutOfRange
	}
	return nil
}

// ValidateUserID は VR-B-06 (userID 必須) を検証する。
//
// 空文字列の場合は ErrInvalidInput を返す。空白のみも空とみなす。
func ValidateUserID(userID string) error {
	if strings.TrimSpace(userID) == "" {
		return ErrInvalidInput
	}
	return nil
}

// ValidateDeductInput は VR-B-03 (amount > 0) + VR-B-04 (key 形式) +
// VR-B-05 (key の userID 一致) を検証する。
//
// 違反時は ErrInvalidInput を返す。
func ValidateDeductInput(userID string, amount int, idempotencyKey string) error {
	if err := ValidateUserID(userID); err != nil {
		return err
	}
	if amount <= 0 {
		return ErrInvalidInput
	}
	if !idempotencyKeyPattern.MatchString(idempotencyKey) {
		return ErrInvalidInput
	}
	prefix := strings.SplitN(idempotencyKey, ":", 2)[0]
	if prefix != userID {
		return ErrInvalidInput
	}
	return nil
}
