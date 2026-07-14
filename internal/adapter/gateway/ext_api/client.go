package extapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/M1shoZ/kvant_test/internal/domain"
)

// Представляет структуру данных внешнего API
// Нужен только для парсинга JSON
type ExternalUserDTO struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Address  struct {
		City string `json:"city"`
	} `json:"address"`
	Company struct {
		Name string `json:"name"`
	} `json:"company"`
}

// для взаимодействия с внешним API
type Client struct {
	httpClient *http.Client
	baseUrl    string
	c          chan struct{} // канал выступает в роли семафора
}

func NewClient(baseUrl string, timeout time.Duration, max int) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: timeout,
		},
		baseUrl: baseUrl,
		c:       make(chan struct{}, max),
	}
}

func (c *Client) GetUsers(ctx context.Context) ([]domain.User, error) {

	// занимаем место в канале
	// если лимит исчерапан выполнение заблокируется пока другой запрос не завершится
	select {
	case c.c <- struct{}{}:
		defer func() { <-c.c }()
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("Ошибка при создании запроса: %w", err)
	}

	// Маскируем запрос под обычный браузер
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Ошибка при запросе к внешнему api: %w", err)
	}
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("api вернул статус: %d", resp.StatusCode)
	}

	var dtos []ExternalUserDTO
	if err := json.NewDecoder(resp.Body).Decode(&dtos); err != nil {
		return nil, fmt.Errorf("Ошибка при декодировании ответа: %w", err)
	}

	users := make([]domain.User, len(dtos))
	for i, dto := range dtos {
		users[i] = domain.User{
			ID:          dto.ID,
			FullName:    dto.Name,
			Username:    dto.Username,
			Email:       dto.Email,
			City:        dto.Address.City,
			CompanyName: dto.Company.Name,
		}
	}
	return users, nil
}

func (c *Client) GetUsersById(ctx context.Context, id int) (*domain.User, error) {
	// занимаем место в канале
	// если лимит исчерапан выполнение заблокируется пока другой запрос не завершится
	select {
	case c.c <- struct{}{}:
		defer func() { <-c.c }()
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	url := fmt.Sprintf("%s/%d", c.baseUrl, id)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("Ошибка при создании запроса: %w", err)
	}

	// Маскируем запрос под обычный браузер
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Ошибка при запросе к внешнему api: %w", err)
	}
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil // если пользователь не найден, возвращаем nil без ошибки
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("api вернул статус: %d", resp.StatusCode)
	}

	var dto ExternalUserDTO
	if err := json.NewDecoder(resp.Body).Decode(&dto); err != nil {
		return nil, fmt.Errorf("Ошибка при декодировании ответа: %w", err)
	}

	return &domain.User{
		ID:          dto.ID,
		FullName:    dto.Name,
		Username:    dto.Username,
		Email:       dto.Email,
		City:        dto.Address.City,
		CompanyName: dto.Company.Name,
	}, nil
}
