package database

import (
	c "domain0/config"
	m "domain0/models"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"time"
)

var DB *gorm.DB

func migrate(db *gorm.DB) error {
	flag := false
	flag = db.SetupJoinTable(&m.User{}, "Domains", &m.UserDomain{}) != nil || flag
	flag = db.AutoMigrate(m.Domain{}) != nil || flag
	flag = db.AutoMigrate(m.DomainChange{}) != nil || flag
	flag = db.AutoMigrate(m.UserDomain{}) != nil || flag
	flag = db.AutoMigrate(m.User{}) != nil || flag
	flag = db.AutoMigrate(m.SSOState{}) != nil || flag
	if flag {
		logrus.Errorf("migrate error")
		return gorm.ErrInvalidDB
	}

	// Clean up soft-deleted rows after removing DeletedAt fields
	if err := cleanupSoftDeletedRows(db); err != nil {
		logrus.Errorf("failed to cleanup soft-deleted rows: %v", err)
		return err
	}

	return nil
}

func cleanupSoftDeletedRows(db *gorm.DB) error {
	logrus.Info("Starting cleanup of soft-deleted rows...")

	// Check if deleted_at columns exist before attempting cleanup
	if hasColumn(db, "domains", "deleted_at") {
		// Delete soft-deleted domains
		result := db.Exec("DELETE FROM domains WHERE deleted_at IS NOT NULL")
		if result.Error != nil {
			logrus.Errorf("failed to delete soft-deleted domains: %v", result.Error)
			return result.Error
		}
		logrus.Infof("Deleted %d soft-deleted domains", result.RowsAffected)

		// Drop deleted_at column from domains table
		if err := db.Migrator().DropColumn(&m.Domain{}, "deleted_at"); err != nil {
			logrus.Warnf("failed to drop deleted_at column from domains: %v", err)
		}
	}

	if hasColumn(db, "domain_changes", "deleted_at") {
		// Delete soft-deleted domain changes
		result := db.Exec("DELETE FROM domain_changes WHERE deleted_at IS NOT NULL")
		if result.Error != nil {
			logrus.Errorf("failed to delete soft-deleted domain changes: %v", result.Error)
			return result.Error
		}
		logrus.Infof("Deleted %d soft-deleted domain changes", result.RowsAffected)

		// Drop deleted_at column from domain_changes table
		if err := db.Migrator().DropColumn(&m.DomainChange{}, "deleted_at"); err != nil {
			logrus.Warnf("failed to drop deleted_at column from domain_changes: %v", err)
		}
	}

	if hasColumn(db, "users", "deleted_at") {
		// Delete soft-deleted users
		result := db.Exec("DELETE FROM users WHERE deleted_at IS NOT NULL")
		if result.Error != nil {
			logrus.Errorf("failed to delete soft-deleted users: %v", result.Error)
			return result.Error
		}
		logrus.Infof("Deleted %d soft-deleted users", result.RowsAffected)

		// Drop deleted_at column from users table
		if err := db.Migrator().DropColumn(&m.User{}, "deleted_at"); err != nil {
			logrus.Warnf("failed to drop deleted_at column from users: %v", err)
		}
	}

	logrus.Info("Completed cleanup of soft-deleted rows")
	return nil
}

func hasColumn(db *gorm.DB, tableName, columnName string) bool {
	return db.Migrator().HasColumn(tableName, columnName)
}

func Init() error {
	var err error
	switch c.CONFIG.Database.Type {
	case "sqlite":
		DB, err = sqliteInit(c.CONFIG)
		if err != nil {
			return err
		}
	case "postgresql":
		DB, err = postgresqlInit(c.CONFIG)
		if err != nil {
			return err
		}
	default:
		logrus.Errorf("database type not supported")
		return gorm.ErrInvalidDB
	}
	go startSSOStateCleaner()
	return nil
}

func startSSOStateCleaner() {
	for {
		time.Sleep(time.Duration(60 * time.Second))
		DB.Where("expired_time < ?", time.Now()).Delete(&m.SSOState{})
	}
}
