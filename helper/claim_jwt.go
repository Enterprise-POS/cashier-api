package common

import (
	"fmt"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

func ClaimJWT(value string, claims *jwt.MapClaims) (*jwt.Token, error) {
	jwtS := os.Getenv("JWT_S")
	if jwtS == "" {
		panic("Environment variable for JWT_S is not defined")
	}

	token, err := jwt.ParseWithClaims(value, claims, func(token *jwt.Token) (interface{}, error) {
		// Optional: verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtS), nil
	})

	return token, err
}
