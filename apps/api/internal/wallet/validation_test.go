package wallet

import (
	"errors"
	"testing"
)

func TestValidateMonthlyBudget(t *testing.T) {
	tests := []struct {
		name    string
		input   int
		wantErr error
	}{
		{"min boundary OK", 1_000, nil},
		{"max boundary OK", 100_000, nil},
		{"typical OK", 30_000, nil},
		{"below min", 999, ErrBudgetOutOfRange},
		{"zero", 0, ErrBudgetOutOfRange},
		{"negative", -1, ErrBudgetOutOfRange},
		{"above max", 100_001, ErrBudgetOutOfRange},
		{"not multiple of 1000", 1_500, ErrBudgetOutOfRange},
		{"not multiple of 1000 (30001)", 30_001, ErrBudgetOutOfRange},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateMonthlyBudget(tc.input)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("got %v, want %v", err, tc.wantErr)
			}
		})
	}
}

func TestValidateDeductInput(t *testing.T) {
	const userA = "user_a"
	const validKey = "user_a:01HW3XKQ7EZ8XN1MX1C7Q4Z9TY"
	tests := []struct {
		name    string
		userID  string
		amount  int
		key     string
		wantErr error
	}{
		{"valid", userA, 100, validKey, nil},
		{"empty userID", "", 100, validKey, ErrInvalidInput},
		{"whitespace userID", "  ", 100, validKey, ErrInvalidInput},
		{"zero amount", userA, 0, validKey, ErrInvalidInput},
		{"negative amount", userA, -1, validKey, ErrInvalidInput},
		{"invalid key format (no colon)", userA, 100, "noKey", ErrInvalidInput},
		{"invalid key format (bad ULID)", userA, 100, "user_a:tooshort", ErrInvalidInput},
		{"key prefix mismatch", "user_b", 100, validKey, ErrInvalidInput},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateDeductInput(tc.userID, tc.amount, tc.key)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("got %v, want %v", err, tc.wantErr)
			}
		})
	}
}

func TestValidateUserID(t *testing.T) {
	if err := ValidateUserID("user_a"); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if err := ValidateUserID(""); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestSentinelErrorsAreDistinct(t *testing.T) {
	if errors.Is(ErrInsufficientBalance, ErrIdempotencyConflict) {
		t.Fatal("ErrInsufficientBalance and ErrIdempotencyConflict must be distinct")
	}
	if errors.Is(ErrBudgetOutOfRange, ErrInvalidInput) {
		t.Fatal("ErrBudgetOutOfRange and ErrInvalidInput must be distinct")
	}
}
