package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"websystem-backend/db"
	"websystem-backend/utils"
)

type CreateUserRequest struct {
	Username     string `json:"username"`
	Password     string `json:"password"`
	Email        string `json:"email"`
	AssignedShop string `json:"assignedShop,omitempty"`
}

type CreateUserResponse struct {
	UserID       string `json:"userId"`
	Username     string `json:"username"`
	Email        string `json:"email"`
	AssignedShop string `json:"assignedShop,omitempty"`
}

func HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	fmt.Println("[INFO] Got Create User Request")

	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" || req.Email == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	var userID string
	err = db.Conn.QueryRow(
		context.Background(),
		`INSERT INTO users (username, password_hash, email, assigned_shop_id) 
         VALUES ($1, $2, $3, $4) RETURNING user_id`,
		req.Username, hashedPassword, req.Email, req.AssignedShop,
	).Scan(&userID)

	if err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	response := CreateUserResponse{
		UserID:       userID,
		Username:     req.Username,
		Email:        req.Email,
		AssignedShop: req.AssignedShop,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
