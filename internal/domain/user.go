package domain

import (
	"context"
	"time"
)

type User struct {
	ID          int    `json:"id"`
	FullName    string `json:"full_name"`
	Username    string `json:"username"`
	Email       string `json:"email"`
	City        string `json:"city"`
	CompanyName string `json:"company_name"`
}

// Получение данных из внешнего апи
type ExternalAPI interface {
	GetUsers(ctx context.Context) ([]User, error)
	GetUsersById(ctx context.Context, id int) (*User, error)
}

// Кэширование данных
type UserCache interface {
	// Получаем значение из кэша по ключу
	Get(ctx context.Context, key string) (any, bool)
	// Сохраняем значение в кэш
	Set(ctx context.Context, key string, value any, ttl time.Duration)
}
