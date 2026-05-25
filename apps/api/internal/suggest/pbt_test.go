package suggest

import (
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	orderhistory "github.com/ryotinjpn/goro2pay/apps/api/internal/repo/order_history"
)

// TestPBT_WithinWindow は P-SG-PBT-01 (NFRD-D12) の軽量プロパティ。
//
// 任意の履歴 (orderedAt をランダムな時間前) に対し、withinWindow は
//   - 入力件数を超えない
//   - 返す全件が cutoff より後 (window 内)
//
// を常に満たす。
func TestPBT_WithinWindow(t *testing.T) {
	props := gopter.NewProperties(nil)
	now := time.Date(2026, 5, 25, 12, 0, 0, 0, time.UTC)
	window := 30 * 24 * time.Hour
	cutoff := now.Add(-window)

	props.Property("withinWindow は window 内のレコードのみ返す", prop.ForAll(
		func(hoursAgo []int) bool {
			recs := make([]orderhistory.OrderRecord, len(hoursAgo))
			for i, h := range hoursAgo {
				recs[i] = orderhistory.OrderRecord{
					Category:  "food",
					OrderedAt: now.Add(-time.Duration(h) * time.Hour).Format(time.RFC3339),
				}
			}
			got := withinWindow(recs, now, window)
			if len(got) > len(recs) {
				return false
			}
			for _, r := range got {
				ts, err := time.Parse(time.RFC3339, r.OrderedAt)
				if err != nil || !ts.After(cutoff) {
					return false
				}
			}
			return true
		},
		gen.SliceOf(gen.IntRange(0, 2400)), // 0〜100 日前
	))

	props.TestingRun(t)
}

// TestPBT_HasRecentSameCategoryOrder は抑制判定 (BR-D03) の単調性プロパティ。
//
// window 内に同カテゴリ注文が 1 件でもあれば必ず true、皆無なら false。
func TestPBT_HasRecentSameCategoryOrder(t *testing.T) {
	props := gopter.NewProperties(nil)
	now := time.Date(2026, 5, 25, 12, 0, 0, 0, time.UTC)
	window := 3 * time.Hour

	props.Property("3h 内同カテゴリ注文の有無を正しく判定する", prop.ForAll(
		func(hoursAgo []int) bool {
			recs := make([]orderhistory.OrderRecord, len(hoursAgo))
			expect := false
			for i, h := range hoursAgo {
				recs[i] = orderhistory.OrderRecord{
					Category:  "food",
					OrderedAt: now.Add(-time.Duration(h) * time.Hour).Format(time.RFC3339),
				}
				if time.Duration(h)*time.Hour < window {
					expect = true
				}
			}
			return hasRecentSameCategoryOrder(recs, now, window, "food") == expect
		},
		gen.SliceOf(gen.IntRange(0, 48)),
	))

	props.TestingRun(t)
}
