package gcaptcha

import (
	"context"
	"sync"
	"time"
)

type memItem struct {
	data    ValidData
	expires time.Time
}

// MemoryStore is an in-memory Store for tests and single-process deployments.
type MemoryStore struct {
	mu    sync.Mutex
	items map[string]memItem
	now   func() time.Time
}

// NewMemoryStore creates an empty in-memory store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{items: map[string]memItem{}, now: time.Now}
}

// SetClock overrides the clock (tests).
func (m *MemoryStore) SetClock(now func() time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if now != nil {
		m.now = now
	}
}

// Save implements Store.
func (m *MemoryStore) Save(_ context.Context, id string, data ValidData, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.items[id] = memItem{data: data, expires: m.now().Add(ttl)}
	return nil
}

// Take implements Store: the entry is removed even when it has expired.
func (m *MemoryStore) Take(_ context.Context, id string) (ValidData, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	item, ok := m.items[id]
	if !ok {
		return ValidData{}, false, nil
	}
	delete(m.items, id)
	if !m.now().Before(item.expires) {
		return ValidData{}, false, nil
	}
	return item.data, true, nil
}
