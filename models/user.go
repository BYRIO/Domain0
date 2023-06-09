package models

import (
	"database/sql"

	"gorm.io/gorm"
)

type UserRole int

type User struct {
	gorm.Model
	Email    string         `gorm:"uniqueIndex"`
	Password string         `gorm:"not null" json:"-"`
	StuId    sql.NullString `gorm:"uniqueIndex"`
	Name     string
	Role     UserRole  `gorm:"default:0"`
	Domains  []*Domain `gorm:"many2many:user_domains;"`
}

func (u *User) GetStuId() *string {
	if !u.StuId.Valid {
		return nil
	}
	return &u.StuId.String
}

const (
	Normal      UserRole = iota // only can access granted domains
	Contributor                 // can submit new domain, access and delete own domain, grant/deny own domain access to other Normal user
	Admin                       // can submit new domain, access and delete all domains, promte/demote user to Contributor, grant/deny all domains access to other Normal user
	SysAdmin                    // same as Admin, promte/demote user to Admin
)

func (u *UserRole) String() string {
	return [...]string{"Normal", "Contributor", "Admin", "SysAdmin"}[*u]
}
