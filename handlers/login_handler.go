package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"websystem-backend/db"
	"websystem-backend/models"
	"websystem-backend/utils"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	UserID       string `json:"userId"`
	Username     string `json:"username"`
	Email        string `json:"email"`
	AssignedShop string `json:"assignedShop,omitempty"`
	Token        string `json:"token"` // トークン (JWTを将来的に使うかも)
}

func HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var user models.User
	err := db.Conn.QueryRow(
		context.Background(),
		"SELECT user_id, username, password_hash, email, assigned_shop_id FROM users WHERE username = $1",
		req.Username,
	).Scan(&user.UserID, &user.Username, &user.PasswordHash, &user.Email, &user.AssignedShop)

	if err != nil {
		fmt.Println(err)
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	if !utils.CheckPassword(user.PasswordHash, req.Password) {
		fmt.Println("Invalid password")
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	// 仮のトークン (JWT 実装の簡易版)
	token := "fake-jwt-token" // 実際には JWT ライブラリで生成する

	response := LoginResponse{
		UserID:       user.UserID,
		Username:     user.Username,
		Email:        user.Email,
		AssignedShop: user.AssignedShop,
		Token:        token,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
