package utils

// import (
// 	models "crud-app/app/model"
// 	"time"

// 	"github.com/golang-jwt/jwt/v5"
// )

var jwtSecret = []byte("mysupersecretkey_1234567890!@#$%^&")

// func GenerateToken(user models.User) (string, error) {
// 	var alumniIDStr string
// 	if user.AlumniID != nil {
// 		alumniIDStr = user.AlumniID.Hex()
// 	}

// 	claims := models.JWTClaims{
// 		UserID:   user.ID,
// 		Username: user.Username,
// 		Role:     user.Role,
// 		AlumniID: alumniIDStr, 
// 		RegisteredClaims: jwt.RegisteredClaims{
// 			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
// 			IssuedAt:  jwt.NewNumericDate(time.Now()),
// 		},
// 	}

// 	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
// 	return token.SignedString(jwtSecret)
// }

// func ValidateToken(tokenString string) (*models.JWTClaims, error) {
// 	token, err := jwt.ParseWithClaims(tokenString, &models.JWTClaims{},
// 		func(token *jwt.Token) (interface{}, error) {
// 			return jwtSecret, nil
// 		},
// 	)

// 	if err != nil {
// 		return nil, err
// 	}

// 	if claims, ok := token.Claims.(*models.JWTClaims); ok && token.Valid {
// 		return claims, nil
// 	}

// 	return nil, jwt.ErrInvalidKey
// }
