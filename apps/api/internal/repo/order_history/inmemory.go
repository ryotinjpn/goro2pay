package orderhistory

import (
	"context"
	"errors"
	"sort"
	"sync"
	"time"
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

// ListRecent は OrderHistoryReader.ListRecent の in-memory 実装。
func (r *InmemoryRepository) ListRecent(ctx context.Context, userID string, limit int) ([]OrderRecord, error) {
	ptrs, err := r.Query(ctx, userID, limit)
	if err != nil {
		return nil, err
	}
	result := make([]OrderRecord, len(ptrs))
	for i, p := range ptrs {
		result[i] = *p
	}
	return result, nil
}

// CountThisMonth は当月 JST 内の注文件数を返す in-memory 実装。
func (r *InmemoryRepository) CountThisMonth(ctx context.Context, userID string) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	monthStart := thisMonthStartUTC()
	count := 0
	for _, rec := range r.byUser[userID] {
		t, err := time.Parse(time.RFC3339, rec.OrderedAt)
		if err != nil {
			continue
		}
		if !t.Before(monthStart) {
			count++
		}
	}
	return count, nil
}

// SumThisMonth は当月 JST 内の注文合計金額を返す in-memory 実装。
func (r *InmemoryRepository) SumThisMonth(ctx context.Context, userID string) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	monthStart := thisMonthStartUTC()
	total := 0
	for _, rec := range r.byUser[userID] {
		t, err := time.Parse(time.RFC3339, rec.OrderedAt)
		if err != nil {
			continue
		}
		if !t.Before(monthStart) {
			total += rec.Amount
		}
	}
	return total, nil
}
