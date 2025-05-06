package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pedersandvoll/Practice-Exam-BE/handlers"
	"github.com/pedersandvoll/Practice-Exam-BE/middleware"
)

func Routes(app *fiber.App, h *handlers.Handlers) {
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, GORM with PostgreSQL!")
	})
	app.Post("/register", h.RegisterUser)
	app.Post("/login", h.LoginUser)

	api := app.Group("/api")
	api.Use(middleware.AuthRequired(h.JWTSecret))
	api.Post("/customers/create", h.RegisterCustomer)
	api.Get("/customers", h.GetCustomers)
}
