package auth

import (
	"strings"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/stretchr/testify/assert"
)

// PBT (Property-Based Testing) - A-NFR-TEST-02 / P-TEST-01
// gopter のデフォルト Generation 数 100、5 プロパティで合計 1-3 秒目安。

// TestPropertyNormalizeIdempotent: Normalize はべき等であること
//
//	Normalize(Normalize(x)) == Normalize(x)
func TestPropertyNormalizeIdempotent(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("Normalize is idempotent", prop.ForAll(
		func(s string) bool {
			once := Normalize(s)
			twice := Normalize(once)
			return once == twice
		},
		gen.AnyString(),
	))

	properties.TestingRun(t)
}

// TestPropertyNormalizeCaseInsensitive: Normalize は大文字小文字不問であること
//
//	Normalize(ToUpper(x)) == Normalize(x)
func TestPropertyNormalizeCaseInsensitive(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("Normalize is case-insensitive", prop.ForAll(
		func(s string) bool {
			return Normalize(strings.ToUpper(s)) == Normalize(s)
		},
		gen.AnyString(),
	))

	properties.TestingRun(t)
}

// TestPropertyHashLength: Hash の出力は常に 64 hex chars (SHA256 hex)
func TestPropertyHashLength(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("Hash output is 64 hex chars", prop.ForAll(
		func(s string) bool {
			h := Hash(s)
			if len(h) != 64 {
				return false
			}
			// hex chars only
			for _, r := range h {
				if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f')) {
					return false
				}
			}
			return true
		},
		gen.AnyString(),
	))

	properties.TestingRun(t)
}

// TestPropertyHashNormalizeConsistency: Hash は正規化済み入力と整合する
//
//	Hash(x) == Hash(Normalize(x))
func TestPropertyHashNormalizeConsistency(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("Hash(x) == Hash(Normalize(x))", prop.ForAll(
		func(s string) bool {
			return Hash(s) == Hash(Normalize(s))
		},
		gen.AnyString(),
	))

	properties.TestingRun(t)
}

// TestPropertyHashCollisionResistance: 異なる入力 (正規化後に異なる) は異なる hash を返す
//
//	Normalize(x) != Normalize(y) ならば Hash(x) != Hash(y)
func TestPropertyHashCollisionResistance(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("distinct normalized inputs yield distinct hashes", prop.ForAll(
		func(x, y string) bool {
			if Normalize(x) == Normalize(y) {
				// 正規化後に同一なら hash も同一でよい (衝突ではない)
				return Hash(x) == Hash(y)
			}
			return Hash(x) != Hash(y)
		},
		gen.AnyString(),
		gen.AnyString(),
	))

	properties.TestingRun(t)
}

// --- Example-based tests ---

func TestNormalize_Examples(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"already lowercase", "user@example.com", "user@example.com"},
		{"uppercase domain", "user@EXAMPLE.COM", "user@example.com"},
		{"mixed case", "User@Example.Com", "user@example.com"},
		{"surrounding whitespace", "  user@example.com  ", "user@example.com"},
		{"empty", "", ""},
		{"unicode", "ユーザー@example.com", "ユーザー@example.com"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, Normalize(tc.in))
		})
	}
}

func TestHash_Examples(t *testing.T) {
	// 確定値で SHA256("user@example.com") = b4c9...（事前計算で確認）
	got := Hash("user@example.com")
	assert.Len(t, got, 64)

	// 大文字でも同一 hash
	gotUpper := Hash("USER@EXAMPLE.COM")
	assert.Equal(t, got, gotUpper)

	// 異なる入力で異なる hash
	gotOther := Hash("other@example.com")
	assert.NotEqual(t, got, gotOther)
}
