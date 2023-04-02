package routers

import (
	"github.com/gofiber/fiber/v2"
)

func InitRouter(fiber *fiber.App) {
	// init public router
	r := fiber.Group("/api/v1")
	SetupUserRouterPub(r)

	// init fiber jwt
	SetUpJwtTokenMiddleware(r)

	// init private router
	SetupUserRouter(r)
	SetupDomainRouter(r)
}
