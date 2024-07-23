package memorycache

import (
	"context"
	"encoding/json"
	"sync"
	"time"
)

type item struct {
	value      []byte
	expiration int64
}

type MemoryCache struct {
	ttl   time.Duration
	items map[string]item
	mu    sync.RWMutex
}

func New(ttl time.Duration) *MemoryCache {
	return &MemoryCache{
		ttl:   ttl,
		items: make(map[string]item),
	}
}

func (m *MemoryCache) Get(ctx context.Context, key string, dest interface{}) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	it, found := m.items[key]
	if !found || (it.expiration > 0 && it.expiration < time.Now().UnixNano()) {
		return false
	}

	err := json.Unmarshal(it.value, dest)
	if err != nil {
		return false
	}

	return true
}

func (m *MemoryCache) Set(ctx context.Context, key string, value interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	b, err := json.Marshal(value)
	if err != nil {
		return err
	}

	expiration := int64(0)
	if m.ttl > 0 {
		expiration = time.Now().Add(m.ttl).UnixNano()
	}

	m.items[key] = item{
		value:      b,
		expiration: expiration,
	}
	return nil
}
