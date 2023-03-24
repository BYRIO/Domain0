package routers

import (
	"domain0/services"

	"github.com/gofiber/fiber/v2"
)

func SetupUserRouter(r fiber.Router) {
	r.Get("/user", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})
	user := r.Group("/user")
	user.Post("/login", services.Login)
	user.Post("/register", services.Register)
}
