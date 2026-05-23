package orderhistory

import (
	"context"
	"errors"
	"sort"
	"sync"
)

// ErrInmemoryDuplicate は InmemoryRepository.Insert で同一 PK+SK が衝突した場合に返す。
var ErrInmemoryDuplicate = errors.New("inmemory: duplicate insert")

// InmemoryRepository は OrderHistoryRepository の in-memory 実装 (LC-20)。
//
// PBT (P-PBT-01) で状態遷移を再現可能にするため、production の DynamoDB と
// 同じ interface を実装する。テスト専用。
type InmemoryRepository struct {
	mu      sync.Mutex
	records map[string]*OrderRecord // key: PK+SK
	byUser  map[string][]*OrderRecord
}

// NewInmemoryRepository は空の InmemoryRepository を返す。
func NewInmemoryRepository() *InmemoryRepository {
	return &InmemoryRepository{
		records: make(map[string]*OrderRecord),
		byUser:  make(map[string][]*OrderRecord),
	}
}

func keyOf(r *OrderRecord) string {
	return pk(r.UserID) + "|" + sk(r.OrderedAt, r.OrderID)
}

// Insert は records map に追加する。同一 key 衝突時は ErrInmemoryDuplicate。
func (r *InmemoryRepository) Insert(ctx context.Context, record *OrderRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	k := keyOf(record)
	if _, exists := r.records[k]; exists {
		return ErrInmemoryDuplicate
	}
	cp := *record
	r.records[k] = &cp
	r.byUser[record.UserID] = append(r.byUser[record.UserID], &cp)
	return nil
}

// GetItem は records map からキー指定で取得する。
func (r *InmemoryRepository) GetItem(ctx context.Context, userID, orderID, orderedAt string) (*OrderRecord, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	k := pk(userID) + "|" + sk(orderedAt, orderID)
	rec, ok := r.records[k]
	if !ok {
		return nil, nil
	}
	cp := *rec
	return &cp, nil
}

// Query は userID の records を orderedAt 降順で limit 件返す。
func (r *InmemoryRepository) Query(ctx context.Context, userID string, limit int) ([]*OrderRecord, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if limit <= 0 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	src := r.byUser[userID]
	dst := make([]*OrderRecord, len(src))
	copy(dst, src)
	sort.Slice(dst, func(i, j int) bool {
		return dst[i].OrderedAt > dst[j].OrderedAt
	})
	if len(dst) > limit {
		dst = dst[:limit]
	}
	return dst, nil
}

// Len はテスト assertion 用にレコード総数を返す。
func (r *InmemoryRepository) Len() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.records)
}
