package database

import (
	"domain0/config"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func sqliteInit(c config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(c.Database.Host), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		return nil, err
	}
	err = migrate(db)
	if err != nil {
		return nil, err
	}
	return db, nil
}
