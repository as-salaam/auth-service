package models

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Credentials struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type Claims struct {
	UserID    uuid.UUID
	UserEmail string
	jwt.RegisteredClaims
}

// todo: move JWT Key to env variables
var JWTKey = []byte("how do you think, what is it?")

const AuthTokenCookieName = "X-Auth-Token"
const AuthTokenCookieTTl = 10080
