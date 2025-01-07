package models

type User struct {
	UserID       int    `json:"userId"`
	Username     string `json:"username"`
	PasswordHash string `json:"-"`
	Email        string `json:"email"`
	AssignedShop string `json:"assignedShop,omitempty"`
}
