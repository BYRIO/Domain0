package models

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

type DomainAction int
type ActionStatus int
type UserDomianRole int

type Domain struct {
	gorm.Model
	Name      string `gorm:"uniqueIndex"`
	ApiId     string `json:"api_id"`
	ApiSecret string `json:"api_secret"`
	Vendor    string `json:"vendor"`
	ICPReg    bool   `json:"ICP_reg" gorm:"column:ICP_reg,default:false"`
	Users     []User `gorm:"many2many:user_domains;"`
}

type DomainChange struct {
	gorm.Model
	Domain       Domain
	User         User
	ActionType   DomainAction // 0: submit, 1: edit DNS, 2: edit others, 3: grant access, 4: revoke access, 5: delete
	ActionStatus ActionStatus // 0: reviewing, 1: approved, 2: rejected
	Reason       string
	Operation    string // json string, describe the operation details
}

type UserDomian struct {
	UserID    uint           `gorm:"primaryKey"`
	DomainID  uint           `gorm:"primaryKey"`
	Role      UserDomianRole // 0: read only, 1: read write, 2: manager, 3: owner
	CreatedAt time.Time
	UpdatedAt time.Time
}

const (
	Submit DomainAction = iota
	EditDNS
	EditOthers
	GrantAccess
	RevokeAccess
	Delete
)

const (
	Reviewing ActionStatus = iota
	Approved
	Rejected
)

const (
	ReadOnly UserDomianRole = iota
	ReadWrite
	Manager
	Owner
)

func (d *Domain) ExtractAuth() (string, string, error) {
	if len(d.ApiId) == 0 || len(d.ApiSecret) == 0 {
		return "", "", errors.New("api id or secret is empty")
	}
	return d.ApiId, d.ApiSecret, nil
}
