package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/M1shoZ/kvant_test/internal/domain"
	"golang.org/x/sync/singleflight"
)

type UserUseCase struct {
	apiClient domain.ExternalAPI
	cache     domain.UserCache
	sfGroup   singleflight.Group
	cacheTTL  time.Duration
}

func NewUserUseCase(apiClient domain.ExternalAPI, cache domain.UserCache, cacheTTL time.Duration) *UserUseCase {
	return &UserUseCase{
		apiClient: apiClient,
		cache:     cache,
		cacheTTL:  cacheTTL,
	}
}

// GetUsers возвращает список пользователей. Сначала проверяется кэш.
// Если данных нет, запрос дедуплицируем и отправляем в апи
func (uc *UserUseCase) GetUsers(ctx context.Context) ([]domain.User, error) {
	cacheKey := "users_all"

	if val, found := uc.cache.Get(ctx, cacheKey); found {
		if users, ok := val.([]domain.User); ok {
			return users, nil
		}
	}

	val, err, _ := uc.sfGroup.Do(cacheKey, func() (any, error) {
		users, err := uc.apiClient.GetUsers(ctx)
		if err != nil {
			return nil, err
		}

		uc.cache.Set(ctx, cacheKey, users, uc.cacheTTL)
		return users, nil
	})

	if err != nil {
		return nil, err
	}

	return val.([]domain.User), nil
}

// GetUserByID возвращает пользователя по его ID. Сначала проверяется кэш.
func (uc *UserUseCase) GetUserByID(ctx context.Context, id int) (*domain.User, error) {
	cacheKey := fmt.Sprintf("user_%d", id)

	val, found := uc.cache.Get(ctx, cacheKey)
	if found {
		if user, ok := val.(*domain.User); ok {
			return user, nil
		}

	}

	val, err, _ := uc.sfGroup.Do(cacheKey, func() (any, error) {
		user, err := uc.apiClient.GetUsersById(ctx, id)
		if err != nil {
			return nil, err
		}

		// если юзер найден, то заносим его в кэш
		if user != nil {
			uc.cache.Set(ctx, cacheKey, user, uc.cacheTTL)
		}
		return user, nil
	})

	if err != nil {
		return nil, err
	}

	return val.(*domain.User), nil
}
