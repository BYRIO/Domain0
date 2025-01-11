package database

import (
	"domain0/config"
	"fmt"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"os"
)

func postgresqlInit(c config.Config) (*gorm.DB, error) {
	timezone, exist := os.LookupEnv("TZ")
	if !exist {
		timezone = "UTC"
	}
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d TimeZone=%s",
		c.Database.Host,
		c.Database.Username,
		c.Database.Password,
		c.Database.DbName,
		c.Database.Port,
		timezone,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		logrus.Errorf("failed to connect to PostgreSQL: %v", err)
		return nil, err
	}

	err = migrate(db)
	if err != nil {
		logrus.Errorf("migration failed: %v", err)
		return nil, err
	}

	return db, nil
}
