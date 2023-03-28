package web

import "domain0/models"

type Domain struct {
	Status int         `json:"status"`
	Errors string      `json:"errors,omitempty"`
	Data   interface{} `json:"data,omitempty"`
}

type DomainInfoUpdate struct {
	Name      *string `json:"name"`
	ApiId     *string `json:"api_id"`
	ApiSecret *string `json:"api_secret"`
	Vendor    *string `json:"vendor"`
	ICPReg    *bool   `json:"ICP_reg"`
}

type DomainUser struct {
	UserId int                   `json:"user_id"`
	Role   models.UserDomainRole `json:"role"`
}

type DomainUserDetail struct {
	UserId     int                   `json:"user_id"`
	Username   string                `json:"username"`
	Email      string                `json:"email"`
	Role       models.UserDomainRole `json:"role"`
	DomainId   int                   `json:"domain_id"`
	DomainName string                `json:"domain_name"`
}
