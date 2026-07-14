package usecase_test

import (
	"context"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/M1shoZ/kvant_test/internal/domain"
	"github.com/M1shoZ/kvant_test/internal/usecase"
)

// MockExternalAPI имитирует поведение внешнего API
type MockExternalApi struct {
	mu               sync.Mutex
	GetUsersFunc     func(ctx context.Context) ([]domain.User, error)
	GetUserByIDFunc  func(ctx context.Context, id int) (*domain.User, error)
	GetUsersCalls    int
	GetUserByIDCalls int
}

func (m *MockExternalApi) GetUsers(ctx context.Context) ([]domain.User, error) {
	m.mu.Lock()
	m.GetUsersCalls++
	m.mu.Unlock()
	return m.GetUsersFunc(ctx)
}

func (m *MockExternalApi) GetUsersById(ctx context.Context, id int) (*domain.User, error) {
	m.mu.Lock()
	m.GetUserByIDCalls++
	m.mu.Unlock()
	return m.GetUserByIDFunc(ctx, id)
}

// MockUserCache имитирует поведение кэша
type MockUserCache struct {
	mu       sync.Mutex
	store    map[string]any
	GetCalls int
	SetCalls int
}

func (m *MockUserCache) Get(ctx context.Context, key string) (any, bool) {
	m.mu.Lock()
	m.GetCalls++
	val, found := m.store[key]
	m.mu.Unlock()
	return val, found
}

func (m *MockUserCache) Set(ctx context.Context, key string, value any, ttl time.Duration) {
	m.mu.Lock()
	m.SetCalls++
	m.store[key] = value
	m.mu.Unlock()
}

// Тест 1: кэш пуст
func TestGetUsers_CacheEmpty(t *testing.T) {
	mockAPI := &MockExternalApi{
		GetUsersFunc: func(ctx context.Context) ([]domain.User, error) {
			return []domain.User{
				{
					ID:          1,
					FullName:    "John Doe",
					Username:    "johnd",
					Email:       "john@example.com",
					City:        "Gwenborough",
					CompanyName: "Romaguera-Crona",
				},
			}, nil
		},
	}

	mockCache := &MockUserCache{
		store: make(map[string]any),
	}

	uc := usecase.NewUserUseCase(mockAPI, mockCache, 1*time.Minute)

	users, err := uc.GetUsers(context.Background())
	if err != nil {
		t.Fatalf("Ожидалось отсутствие ошибок, получили: %v", err)
	}

	expected := []domain.User{
		{
			ID:          1,
			FullName:    "John Doe",
			Username:    "johnd",
			Email:       "john@example.com",
			City:        "Gwenborough",
			CompanyName: "Romaguera-Crona",
		},
	}
	if !reflect.DeepEqual(users, expected) {
		t.Errorf("Ожидалось %v, получили %v", expected, users)
	}

	// Проверяем кол-во вызовов методов
	if mockCache.GetCalls != 1 {
		t.Errorf("Ожидался 1 вызов Cache.Get, получили %d", mockCache.GetCalls)
	}
	if mockCache.SetCalls != 1 {
		t.Errorf("Ожидался 1 вызов Cache.Set, получили %d", mockCache.SetCalls)
	}
	if mockAPI.GetUsersCalls != 1 {
		t.Errorf("Ожидался 1 вызов API.GetUsers, получили %d", mockAPI.GetUsersCalls)
	}
}

// Тест 2: в кэше есть данные
func TestGetUsers_HasCache(t *testing.T) {
	mockAPI := &MockExternalApi{
		GetUsersFunc: func(ctx context.Context) ([]domain.User, error) {
			t.Errorf("API не должен быть вызван при заполненном кэше")
			return nil, nil
		},
	}

	// В кэше есть данные
	mockCache := &MockUserCache{
		store: map[string]any{
			"users_all": []domain.User{{
				ID:          1,
				FullName:    "Caches User",
				Username:    "cache",
				Email:       "cache@example.com",
				City:        "cacheCity",
				CompanyName: "Romaguera-Crona"}},
		},
	}

	uc := usecase.NewUserUseCase(mockAPI, mockCache, 1*time.Minute)

	users, err := uc.GetUsers(context.Background())
	if err != nil {
		t.Fatalf("Ожидалось отсутствие ошибок, получили: %v", err)
	}

	expected := []domain.User{{
		ID:          1,
		FullName:    "Caches User",
		Username:    "cache",
		Email:       "cache@example.com",
		City:        "cacheCity",
		CompanyName: "Romaguera-Crona"}}
	if !reflect.DeepEqual(users, expected) {
		t.Errorf("Ожидалось %v, получили %v", expected, users)
	}

	if mockCache.GetCalls != 1 {
		t.Errorf("Ожидался 1 вызов Cache.Get, получили %d", mockCache.GetCalls)
	}
	if mockAPI.GetUsersCalls != 0 {
		t.Errorf("Ожидалось 0 вызовов API.GetUsers, получили %d", mockAPI.GetUsersCalls)
	}
}

// Тест 3: проверка дедубликации
func TestGetUsers_Deduplication(t *testing.T) {
	var apiCalls int
	var mu sync.Mutex

	mockAPI := &MockExternalApi{
		GetUsersFunc: func(ctx context.Context) ([]domain.User, error) {
			mu.Lock()
			apiCalls++
			mu.Unlock()
			time.Sleep(50 * time.Millisecond)
			return []domain.User{{
				ID:          1,
				FullName:    "Concurrent User",
				Username:    "Concurrent",
				Email:       "Concurrent@example.com",
				City:        "ConcurrentCity",
				CompanyName: "Romaguera-Crona"}}, nil
		},
	}

	mockCache := &MockUserCache{
		store: make(map[string]any),
	}

	uc := usecase.NewUserUseCase(mockAPI, mockCache, 1*time.Minute)

	// Запускаем 3 параллельных запроса
	numReq := 3
	var wg sync.WaitGroup
	result := make([][]domain.User, numReq)

	for i := 0; i < numReq; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			users, _ := uc.GetUsers(context.Background())
			result[idx] = users
		}(i)
	}

	wg.Wait()

	for _, res := range result {
		if len(res) != 1 || res[0].FullName != "Concurrent User" {
			t.Errorf("ожидался другой результат от горутины: %v", res)
		}
	}

	if apiCalls != 1 {
		t.Errorf("Ожидалось что апи будет вызвано 1 раз, но было вызвано %d раз", apiCalls)
	}
}
