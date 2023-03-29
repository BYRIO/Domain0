package routers

import (
	"domain0/services"

	"github.com/gofiber/fiber/v2"
)

// Path: routers/user.go
func SetupUserRouterPub(r fiber.Router) {
	r.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})
	user := r.Group("/user")
	user.Post("/login", services.Login)
	user.Post("/register", services.Register)
}

func SetupUserRouter(r fiber.Router) {
	user := r.Group("/user")
	user.Get("/", services.UserList)
	user.Get("/:id", services.UserInfoGet)
	user.Put("/:id", services.UserInfoUpdate)
	user.Delete("/:id", services.UserInfoDelete)
}
