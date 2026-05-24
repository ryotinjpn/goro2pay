package fallback

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildFromHistory_Empty(t *testing.T) {
	p := NewSimpleFallbackProvider()
	assert.Nil(t, p.BuildFromHistory(nil))
	assert.Nil(t, p.BuildFromHistory([]HistoryItem{}))
}

func TestBuildFromHistory_MostFrequent(t *testing.T) {
	p := NewSimpleFallbackProvider()
	history := []HistoryItem{
		{StoreName: "ゴロゴロ食堂", MenuName: "おまかせ定食", Amount: 1000, Category: "food"},
		{StoreName: "ぐうたら亭", MenuName: "手抜き丼", Amount: 800, Category: "food"},
		{StoreName: "ゴロゴロ食堂", MenuName: "おまかせ定食", Amount: 1000, Category: "food"},
		{StoreName: "ゴロゴロ食堂", MenuName: "おまかせ定食", Amount: 1000, Category: "food"},
	}
	plan := p.BuildFromHistory(history)
	require.NotNil(t, plan)
	assert.Equal(t, "ゴロゴロ食堂", plan.StoreName)
	assert.Equal(t, "おまかせ定食", plan.MenuName)
	assert.Equal(t, "fallback_history", plan.Source)
}

func TestBuildFromHistory_SinglePattern(t *testing.T) {
	p := NewSimpleFallbackProvider()
	history := []HistoryItem{
		{StoreName: "ぐうたら亭", MenuName: "手抜き丼", Amount: 800, Category: "food"},
	}
	plan := p.BuildFromHistory(history)
	require.NotNil(t, plan)
	assert.Equal(t, "ぐうたら亭", plan.StoreName)
}

func TestDefault_AlwaysReturnsValidStore(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	p := NewSimpleFallbackProviderWithRand(rng)
	for i := 0; i < 20; i++ {
		plan := p.Default()
		require.NotNil(t, plan)
		assert.Equal(t, "fallback_default", plan.Source)
		assert.NotEmpty(t, plan.StoreName)
		assert.NotEmpty(t, plan.MenuName)
		assert.Equal(t, "food", plan.Category)
		assert.True(t, plan.Amount >= 800 && plan.Amount <= 1500)
	}
}

func TestDefault_DeterministicWithSeed(t *testing.T) {
	rng1 := rand.New(rand.NewSource(123))
	rng2 := rand.New(rand.NewSource(123))
	p1 := NewSimpleFallbackProviderWithRand(rng1)
	p2 := NewSimpleFallbackProviderWithRand(rng2)
	for i := 0; i < 5; i++ {
		assert.Equal(t, p1.Default().StoreName, p2.Default().StoreName)
	}
}

func TestFakeFallbackProvider_DefaultBehavior(t *testing.T) {
	f := &FakeFallbackProvider{}
	p1 := f.BuildFromHistory(nil)
	p2 := f.Default()
	assert.Equal(t, "fake-history", p1.StoreName)
	assert.Equal(t, "fake-default", p2.StoreName)
	assert.Equal(t, 1, f.BuildCalls)
	assert.Equal(t, 1, f.DefaultCalls)
}
