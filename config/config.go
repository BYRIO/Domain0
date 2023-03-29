package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type DatabaseConfig struct {
	Type     string `yaml:"type"` // sqlite, mysql, postgres
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	DbName   string `yaml:"db_name"`
}

type Config struct {
	BindAddr string         `yaml:"bind_addr"`
	Database DatabaseConfig `yaml:"database"`
	LogLevel int            `yaml:"log_level"` // 0: debug, 1: info, 2: warn, 3: error
	JwtKey   string         `yaml:"jwt_key"`
}

var CONFIG = Config{
	BindAddr: "127.0.0.1:8080",
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
