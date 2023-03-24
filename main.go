package main

import (
	"domain0/config"
	"domain0/database"

	"github.com/sirupsen/logrus"
)

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

}
