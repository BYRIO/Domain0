package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type DatabaseConfig struct {
	Type     string
	Host     string
	Port     int
	Username string
	Password string
	DbName   string
}

type Config struct {
	Database DatabaseConfig
	LogLevel int // 0: debug, 1: info, 2: warn, 3: error
	JwtKey   string
}

var CONFIG = Config{
	Database: DatabaseConfig{
		Type:     "sqlite",
		Host:     "./db.sqlite3",
		Port:     0,
		Username: "",
		Password: "",
		DbName:   "",
	},
	LogLevel: 1,
	JwtKey:   "secretissecretbutsecretisnotsecure",
}

func Read(filename string) error {
	yamlFile, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(yamlFile, &CONFIG)
	if err != nil {
		return err
	}
	return nil
}
