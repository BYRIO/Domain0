package models

import "time"

type SSOState struct {
	State       string `gorm:"primaryKey"`
	ExpiredTime time.Time
}
