package main

import (
	"fmt"
	"github.com/as-salaam/auth-service/internal/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func main() {
	fmt.Println(models.Claims{
		UserID:    uuid.UUID{},
		UserEmail: "",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "",
			Subject:   "",
			Audience:  nil,
			ExpiresAt: nil,
			NotBefore: nil,
			IssuedAt:  nil,
			ID:        "",
		},
	})
}
