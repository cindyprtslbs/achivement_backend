package models

import "github.com/golang-jwt/jwt/v5"

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginUser struct {
	ID          string   `json:"id"`
	Username    string   `json:"username"`
	FullName    string   `json:"fullName"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
}

type LoginData struct {
	Token        string    `json:"token"`
	RefreshToken string    `json:"refreshToken"`
	User         LoginUser `json:"user"`
}

type LoginResponse struct {
	Status string    `json:"status"`
	Data   LoginData `json:"data"`
}

type JWTClaims struct {
	UserID      string   `json:"user_id"`
	Username    string   `json:"username"`
	RoleName    string   `json:"role_name"`
	Permissions []string `json:"permissions"`
	jwt.RegisteredClaims
}
