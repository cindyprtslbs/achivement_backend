package models

import "github.com/golang-jwt/jwt/v5"

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token       string   `json:"access_token"`
	Refresh     string   `json:"refresh_token"`
	UserID      string   `json:"user_id"`
	Username    string   `json:"username"`
	FullName    string   `json:"full_name"`
	RoleName    string   `json:"role_name"`
	Permissions []string `json:"permissions"`
}

type JWTClaims struct {
	UserID      string   `json:"user_id"`
	Username    string   `json:"username"`
	RoleName    string   `json:"role_name"`
	Permissions []string `json:"permissions"`
	jwt.RegisteredClaims
}
