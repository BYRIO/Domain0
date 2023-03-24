package database

import (
	c "domain0/config"
	m "domain0/models"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var DB *gorm.DB

func migrate(db *gorm.DB) error {
	flag := false
	flag = db.AutoMigrate(m.Domain{}) != nil || flag
	flag = db.AutoMigrate(m.DomainChange{}) != nil || flag
	flag = db.AutoMigrate(m.UserDomian{}) != nil || flag
	flag = db.AutoMigrate(m.User{}) != nil || flag
	if flag {
		logrus.Errorf("migrate error")
		return gorm.ErrInvalidDB
	}
	return nil
}

func Init() error {
	var err error
	switch c.CONFIG.Database.Type {
	case "sqlite":
		DB, err = sqliteInit(c.CONFIG)
		if err != nil {
			return err
		}
	default:
		logrus.Errorf("database type not supported")
		return gorm.ErrInvalidDB
	}
	return nil
}
