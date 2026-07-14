package cache

import (
	"context"
	"sync"
	"time"
)

// оборачивает данные вместе со временем их истечения
type cacheItem struct {
	value     any
	expiresAt time.Time
}

// проверка, истек ли срок жизни элемента
func (item cacheItem) isExpired() bool {
	return time.Now().After(item.expiresAt)
}

type MemoryCache struct {
	mu    sync.RWMutex
	items map[string]cacheItem
}

func NewMemoryCache(cleanupInterval time.Duration) *MemoryCache {
	mc := &MemoryCache{
		items: make(map[string]cacheItem),
	}
	// запуск фонового уборщика просроченного кэша
	go mc.startJanitor(cleanupInterval)
	return mc
}

func (mc *MemoryCache) Get(ctx context.Context, key string) (any, bool) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	item, found := mc.items[key]
	if !found {
		return nil, false
	}

	if item.isExpired() {
		delete(mc.items, key)
		return nil, false
	}

	return item.value, true
}

func (mc *MemoryCache) Set(ctx context.Context, key string, value any, ttl time.Duration) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.items[key] = cacheItem{
		value:     value,
		expiresAt: time.Now().Add(ttl),
	}
}

func (mc *MemoryCache) startJanitor(interval time.Duration) {
	ticker := time.NewTicker(interval)
	for range ticker.C {
		mc.cleanup()
	}
}

func (mc *MemoryCache) cleanup() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	now := time.Now()
	for k, item := range mc.items {
		if now.After(item.expiresAt) {
			delete(mc.items, k)
		}
	}
}
