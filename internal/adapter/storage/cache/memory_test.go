package cache_test

import (
	"context"
	"testing"
	"time"

	"github.com/M1shoZ/kvant_test/internal/adapter/storage/cache"
)

// Тест проверяет ленивое удаление просроченной записи при обращении
func TestMemoryCache_TTLEnd(t *testing.T) {
	// интервал фоновой очистки ставим большим,
	// чтобы проверить именно "ленивое" удаление
	mc := cache.NewMemoryCache(10 * time.Minute)

	ctx := context.Background()
	key := "test_ttl_key"
	val := "hello_world"
	ttl := 50 * time.Millisecond // короткий TTL для быстрого теста

	mc.Set(ctx, key, val, ttl)

	got, found := mc.Get(ctx, key)
	if !found {
		t.Fatal("ожидалось, что данные будут найдены в кэше сразу после записи")
	}
	if got != val {
		t.Errorf("ожидалось значение %v, получено %v", val, got)
	}

	time.Sleep(70 * time.Millisecond)

	_, foundAfterTTL := mc.Get(ctx, key)
	if foundAfterTTL {
		t.Error("ожидалось, что данные устареют и удалятся по истечении TTL")
	}
}

// Тест проверяет, что фоновый уборщик действительно просыпается
// и удаляет просроченные элементы из памяти без нашего обращения
func TestMemoryCache_JanitorCleanup(t *testing.T) {
	mc := cache.NewMemoryCache(20 * time.Millisecond)

	ctx := context.Background()
	key := "test_janitor_key"
	val := "some_data"
	ttl := 10 * time.Millisecond

	mc.Set(ctx, key, val, ttl)

	// Спим 50 миллисекунд, чтобы фоновый уборщик гарантированно сработал хотя бы один раз
	time.Sleep(50 * time.Millisecond)

	_, found := mc.Get(ctx, key)
	if found {
		t.Error("ожидалось, что фоновый уборщик очистит просроченную запись")
	}
}
