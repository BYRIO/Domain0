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

type FeishuConfig struct {
	Enable      bool   `yaml:"enable"`
	AppID       string `yaml:"app_id"`
	AppSecret   string `yaml:"app_secret"`
	RedirectURL string `yaml:"redirect_url"`
	BotUrl      string `yaml:"bot_url"`
}
type OIDCConfig struct {
	LogoURL     string       `yaml:"logo_url"`
	Name        string       `yaml:"name"`
	Enable      bool         `yaml:"enable"`
	AuthURL     string       `yaml:"auth_url"`
	TokenURL    string       `yaml:"token_url"`
	UserInfoURL string       `yaml:"user_info_url"`
	ClientId    string       `yaml:"client_id"`
	AppSecret   string       `yaml:"app_secret"`
	RedirectUrl string       `yaml:"redirect_url"`
	Scope       string       `yaml:"scope"`
	InfoPath    OIDCInfoPath `yaml:"info_path"`
}
type OIDCInfoPath struct {
	Name  string `yaml:"name"`
	Id    string `yaml:"id"`
	Email string `yaml:"email"`
	Error string `yaml:"error"`
}
type Config struct {
	BindAddr string         `yaml:"bind_addr"`
	Database DatabaseConfig `yaml:"database"`
	LogLevel int            `yaml:"log_level"` // 0: debug, 1: info, 2: warn, 3: error
	JwtKey   string         `yaml:"jwt_key"`
	Feishu   FeishuConfig   `yaml:"feishu"`
	OIDC     OIDCConfig     `yaml:"oidc"`
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
	Feishu: FeishuConfig{
		AppID:       "",
		AppSecret:   "",
		RedirectURL: "",
	},
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
