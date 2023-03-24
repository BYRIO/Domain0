package web

type User struct {
	Status interface{} `json:"status,omitempty"`
	Errors interface{} `json:"error,omitempty"`
	Data   interface{} `json:"data,omitempty"`
}

type UserInfoUpdate struct {
	Email    interface{} `json:"email,omitempty"`
	Password interface{} `json:"password,omitempty"`
	StuId    interface{} `json:"stuid,omitempty"`
	Name     interface{} `json:"name,omitempty"`
	Role     interface{} `json:"role,omitempty"`
}
