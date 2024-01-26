package handlers

import (
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

func VerifyJwt(jwToken string, jwtKey []byte, claims jwt.Claims) error {
	token, err := jwt.ParseWithClaims(jwToken, claims, func(token *jwt.Token) (any, error) {
		return jwtKey, nil
	})
	if err != nil {
		return err
	}
	if !token.Valid {
		return fmt.Errorf("Invalid JWT")
	}
	return nil
}
