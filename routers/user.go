package routers

import (
	"domain0/config"
	"domain0/services"

	"github.com/gofiber/fiber/v2"
	jwtware "github.com/gofiber/jwt/v3"
)

// Path: routers/user.go
func SetupUserRouterPub(r fiber.Router) {
	r.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})
	user := r.Group("/user")
	user.Post("/login", services.Login)
	user.Post("/register", services.Register)
	user.Get("/feishu/enable", services.FeishuAuthEnable)
	user.Get("/feishu", services.FeishuAuthRedirect)
	user.Get("/oidc/enable", services.OIDCAuthEnable)
	user.Get("/oidc", services.OIDCAuthRedirect)
	user.Get("/callback", services.Callback)
}

func SetupUserRouter(r fiber.Router) {
	user := r.Group("/user")
	user.Get("/", services.UserList)
	user.Get("/:id", services.UserInfoGet)
	user.Put("/:id", services.UserInfoUpdate)
	user.Delete("/:id", services.UserInfoDelete)
}

func SetUpJwtTokenMiddleware(r fiber.Router) {
	r.Use(jwtware.New(jwtware.Config{
		SigningKey: []byte(config.CONFIG.JwtKey),
	}))
	r.Use(services.JwtToLocalsWare)
}
