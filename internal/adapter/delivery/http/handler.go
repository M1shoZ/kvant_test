package http

import (
	"log"
	"strconv"

	"github.com/M1shoZ/kvant_test/internal/usecase"
	"github.com/gofiber/fiber/v3"
)

// обрабатывает запросы
type UserHandler struct {
	useCase *usecase.UserUseCase
}

func NewUserHandler(uc *usecase.UserUseCase) *UserHandler {
	return &UserHandler{
		useCase: uc,
	}
}

func (h *UserHandler) GetUsers(c fiber.Ctx) error {
	ctx := c.Context()

	users, err := h.useCase.GetUsers(ctx)
	if err != nil {

		log.Printf("[ERROR] GetUsers failed: %v", err)

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Ошибка при получении пользователей",
		})
	}
	return c.JSON(users)
}

func (h *UserHandler) GetUserByID(c fiber.Ctx) error {
	ctx := c.Context()

	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Некорректный id (должен быть числовой)",
		})
	}

	user, err := h.useCase.GetUserByID(ctx, id)
	if err != nil {

		log.Printf("[ERROR] GetUserByID failed: %v", err)

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Ошибка при получении пользователя по id",
		})
	}

	if user == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Пользователь не найден",
		})
	}

	return c.JSON(user)
}
