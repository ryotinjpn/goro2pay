package budget_raise

import "time"

const (
	maxBudget = 100_000
	minBudget = 1
)

var jstLocation = time.FixedZone("Asia/Tokyo", 9*60*60)

// computeRecommendedBudget は推奨増額後予算を返す純関数 (BR-R01)。
//
// recommended = min(floor(current * 1.5), 100_000)
func computeRecommendedBudget(currentBudget int) int {
	recommended := int(float64(currentBudget) * 1.5)
	if recommended > maxBudget {
		return maxBudget
	}
	return recommended
}

// computeNextMonthStart は JST での翌月 1 日 00:00:00 を返す純関数 (BR-R03)。
func computeNextMonthStart() time.Time {
	now := time.Now().In(jstLocation)
	next := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, jstLocation)
	return next
}
