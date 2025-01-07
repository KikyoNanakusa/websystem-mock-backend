package models

type User struct {
	UserID       string `json:"userId"`
	Username     string `json:"username"`
	PasswordHash string `json:"-"`
	Email        string `json:"email"`
	AssignedShop string `json:"assignedShop,omitempty"`
}
