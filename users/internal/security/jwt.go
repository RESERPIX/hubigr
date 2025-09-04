package security

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID int64  `json:"uid"`
	Role   string `json:"role"`
	Nick   string `json:"nick"`
	jwt.RegisteredClaims
}

func SignJWT(userID int64, role, nick string) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "dev-secret-change-me"
	}
	now := time.Now()
	claims := &Claims{
		UserID: userID, Role: role, Nick: nick,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(30 * 24 * time.Hour)),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString([]byte(secret))
}
