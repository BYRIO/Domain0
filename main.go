package main

import (
	"domain0/config"
	"domain0/database"
	_ "domain0/docs"
	"domain0/routers"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
	"github.com/sirupsen/logrus"
)

// @title Domain0 API
// @description Domain0 API
// @version 0.0.1
// @schemes http https
// @host localhost:8080
// @contact.name domain0
// @contact.email makiras.x@outlook.com
// @license.name MPL(mozilla public license)-2.0
// @license.url https://www.mozilla.org/en-US/MPL/2.0/
func main() {
	// read config file
	if err := config.Read("./config.yaml"); err != nil {
		logrus.Error("Failed to read config file")
		logrus.Fatal(err)
	}

	// init database
	if err := database.Init(); err != nil {
		logrus.Error("Failed to init database")
		logrus.Fatal(err)
	}

	f := fiber.New(fiber.Config{
		// set fiber config
	})

	// init swagger
	f.Get("/swagger/*", swagger.HandlerDefault)

	// init router
	routers.InitRouter(f)

	f.Listen(config.CONFIG.BindAddr)

}
