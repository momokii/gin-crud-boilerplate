package models

type UserModelRes struct {
	Id       int    `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name"`
	Role     int    `json:"role"`
	IsActive bool   `json:"is_active"`
}

type UserModel struct {
	UserModelRes
	Password  string `json:"password"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
