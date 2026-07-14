package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/M1shoZ/kvant_test/internal/adapter/delivery/http"
	extapi "github.com/M1shoZ/kvant_test/internal/adapter/gateway/ext_api"
	"github.com/M1shoZ/kvant_test/internal/adapter/storage/cache"
	"github.com/M1shoZ/kvant_test/internal/usecase"
	"github.com/gofiber/fiber/v3"
)

func main() {

	port := getEnv("PORT", "8080")
	externalAPIURL := getEnv("EXTERNAL_API_URL", "https://jsonplaceholder.typicode.com/users")

	cacheTTL := 1 * time.Minute
	apiTimeout := 10 * time.Second
	cleanupInterval := 2 * time.Minute

	cacheAdapter := cache.NewMemoryCache(cleanupInterval)
	apiClient := extapi.NewClient(externalAPIURL, apiTimeout, 10)

	userUseCase := usecase.NewUserUseCase(apiClient, cacheAdapter, cacheTTL)

	userHandler := http.NewUserHandler(userUseCase)

	app := fiber.New()
	http.RegRoutes(app, userHandler)

	go func() {
		log.Printf("Сервер запущен на порту %s", port)
		if err := app.Listen(":" + port); err != nil {
			log.Printf("Сервер остановлен")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit
	log.Println("Получен сигнал остановки, выполняем graceful shutdown")

	if err := app.Shutdown(); err != nil {
		log.Fatalf("Сервер остановлен с ошибкой: %v", err)
	}

	log.Println("Сервер остановлен успешно!")
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
