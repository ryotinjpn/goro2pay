package suggestion

import (
	"context"
	"sync"
	"time"
)

// InmemoryStore は SuggestionStore の test-only in-memory 実装 (LC-SUGGEST-08)。
//
// TTL 失効を時刻関数で模倣する。production では使わない。
type InmemoryStore struct {
	mu    sync.Mutex
	items map[string]*SuggestionRecord
	now   func() time.Time
}

// NewInmemoryStore は空の InmemoryStore を返す。
func NewInmemoryStore() *InmemoryStore {
	return &InmemoryStore{
		items: make(map[string]*SuggestionRecord),
		now:   time.Now,
	}
}

// SetClock は now() を差し替える (TTL 失効テスト用)。
func (s *InmemoryStore) SetClock(now func() time.Time) { s.now = now }

// Save は record を保存する。
func (s *InmemoryStore) Save(_ context.Context, rec *SuggestionRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	cp := *rec
	s.items[rec.SuggestionID] = &cp
	return nil
}

// Get は suggestionId で取得する。TTL (expiresAt) を過ぎていれば nil (DynamoDB
// TTL 削除を模倣)。
func (s *InmemoryStore) Get(_ context.Context, suggestionID string) (*SuggestionRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	rec, ok := s.items[suggestionID]
	if !ok {
		return nil, nil
	}
	if s.now().Unix() >= rec.ExpiresAt {
		return nil, nil
	}
	cp := *rec
	return &cp, nil
}
