package metrics

import (
	"fmt"

	"github.com/ryotinjpn/goro2pay/apps/api/internal/repo/budget_settings"
	"github.com/ryotinjpn/goro2pay/apps/api/internal/repo/wallet_repo"
)

// computeConsumptionRate は消化率を計算する純関数 (BR-M03)。
//
// monthlyBudget > 0 が呼出元で保証されている前提。
// remainingBalance < 0 の場合は 1.0 にクランプ (防御実装)。
func computeConsumptionRate(monthlyBudget, remainingBalance int) float64 {
	spent := monthlyBudget - remainingBalance
	rate := float64(spent) / float64(monthlyBudget)
	if rate < 0.0 {
		return 0.0
	}
	if rate > 1.0 {
		return 1.0
	}
	return rate
}

// computeMetrics は BudgetSettings / Wallet / DamageCount から Metrics を組み立てる純関数。
func computeMetrics(bs *budget_settings.BudgetSettings, w *wallet_repo.WalletRecord, damageCount int) *Metrics {
	rate := computeConsumptionRate(bs.MonthlyBudget, w.Balance)
	spent := bs.MonthlyBudget - w.Balance
	if spent < 0 {
		spent = 0
	}
	summaryText := fmt.Sprintf("今月のダメ化回数: %d 回、消化額 ¥%s", damageCount, formatYen(spent))
	return &Metrics{
		DamageCount:       damageCount,
		ConsumptionRate:   rate,
		MonthlyBudget:     bs.MonthlyBudget,
		RemainingBalance:  w.Balance,
		ThresholdExceeded: rate > 0.8,
		SummaryText:       summaryText,
	}
}

// formatYen は金額を "1,234" 形式に整形する。
func formatYen(amount int) string {
	s := fmt.Sprintf("%d", amount)
	result := []byte{}
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result = append(result, ',')
		}
		result = append(result, byte(c))
	}
	return string(result)
}
