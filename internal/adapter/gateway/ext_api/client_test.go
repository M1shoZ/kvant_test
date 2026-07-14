package extapi_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	extapi "github.com/M1shoZ/kvant_test/internal/adapter/gateway/ext_api"
)

// Тест на проверку работы ограничения кол-ва одновременных запросов
func TestClient_ConcurrencyLimit(t *testing.T) {
	var mu sync.Mutex
	activeRequests := 0
	maxActiveRequests := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		activeRequests++
		// Фиксируем пиковое количество одновременно обрабатываемых сервером запросов
		if activeRequests > maxActiveRequests {
			maxActiveRequests = activeRequests
		}
		mu.Unlock()

		time.Sleep(30 * time.Millisecond)

		mu.Lock()
		activeRequests--
		mu.Unlock()

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[]`))
	}))
	defer server.Close()

	maxConcurrent := 2
	client := extapi.NewClient(server.URL, 5*time.Second, maxConcurrent)

	// Запускаем 5 параллельных запросов к клиенту
	numRequests := 5
	var wg sync.WaitGroup
	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = client.GetUsers(context.Background())
		}()
	}
	wg.Wait()

	if maxActiveRequests > maxConcurrent {
		t.Errorf("Ожидалось максимум %d одновременных запросов к серверу, зафиксировано: %d", maxConcurrent, maxActiveRequests)
	} else {
		t.Logf("Тест успешно пройден. Пиковая нагрузка на сервер составила ровно %d одновременных запросов.", maxActiveRequests)
	}
}
