package http

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
)

func RegRoutes(app *fiber.App, handler *UserHandler) {
	// Встроенное логирование fiber
	app.Use(logger.New())

	app.Get("/users", handler.GetUsers)
	app.Get("/users/:id", handler.GetUserByID)
}
