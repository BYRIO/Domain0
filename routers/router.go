package routers

import (
	"domain0/config"

	"github.com/gofiber/fiber/v2"
	jwtware "github.com/gofiber/jwt/v3"
)

func InitRouter(fiber *fiber.App) {
	// init public router
	r := fiber.Group("/api/v1")
	SetupUserRouterPub(r)

	// init fiber jwt
	r.Use(jwtware.New(jwtware.Config{
		SigningKey: []byte(config.CONFIG.JwtKey),
	}))

	// init private router
	SetupUserRouter(r)
	SetupDomainRouter(r)
}
