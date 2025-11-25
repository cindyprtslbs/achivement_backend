package utils

import (
	"time"

	models "achievement_backend/app/model"

	"github.com/golang-jwt/jwt/v5"
)

var JwtSecret = []byte("mysupersecretkey_1234567890!@#$%^&")

func GenerateToken(user models.User, permissions []string) (string, error) {
	claims := models.JWTClaims{
		UserID:      user.ID,
		Username:    user.Username,
		RoleID:      user.RoleID,
		Permissions: permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(JwtSecret)
}

func ValidateToken(tokenString string) (*models.JWTClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&models.JWTClaims{},
		func(t *jwt.Token) (interface{}, error) {
			return JwtSecret, nil
		},
	)

	if err != nil {
		return nil, err
	}

	claims := token.Claims.(*models.JWTClaims)
	return claims, nil
}
