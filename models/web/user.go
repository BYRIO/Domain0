package web

import "domain0/models"

type User struct {
	Status interface{} `json:"status,omitempty"`
	Errors interface{} `json:"error,omitempty"`
	Data   interface{} `json:"data,omitempty"`
}

type UserInfoUpdate struct {
	Email    *string          `json:"email,omitempty"`
	Password *string          `json:"password,omitempty"`
	StuId    *string          `json:"stuid,omitempty"`
	Name     *string          `json:"name,omitempty"`
	Role     *models.UserRole `json:"role,omitempty"`
}
