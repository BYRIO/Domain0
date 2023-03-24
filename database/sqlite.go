package database

import (
	"domain0/config"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func sqliteInit(c config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(c.Database.Host), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	err = migrate(db)
	if err != nil {
		return nil, err
	}
	return db, nil
}
