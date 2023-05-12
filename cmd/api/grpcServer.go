package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/as-salaam/auth-service/internal/auth"
	"github.com/as-salaam/auth-service/internal/models"
	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"gorm.io/gorm"
	"log"
	"net"
)

func RunGRPC(port uint, db *gorm.DB) (err error) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return errors.New("cannot create listener: " + err.Error())
	}

	serverRegistrar := grpc.NewServer()
	service := &authServer{
		DB: db,
	}

	auth.RegisterAuthServer(serverRegistrar, service)

	log.Printf("Listening and serving RPC on port :%d\n", port)
	err = serverRegistrar.Serve(lis)
	if err != nil {
		return errors.New("impossible to serve: " + err.Error())
	}

	return nil
}

type authServer struct {
	auth.UnimplementedAuthServer
	DB *gorm.DB
}

func (a authServer) Authenticate(ctx context.Context, req *auth.AuthenticateRequest) (*auth.AuthenticateResponse, error) {
	claims := &models.Claims{}
	token, err := jwt.ParseWithClaims(req.Token, claims, func(token *jwt.Token) (interface{}, error) {
		return models.JWTKey, nil
	})

	if err != nil || !token.Valid {
		log.Printf("token validation failed: %s", err)
		return &auth.AuthenticateResponse{
			Authenticated: false,
			User:          nil,
		}, nil
	}

	var dbUser models.User
	if err = a.DB.Where("id = ?", claims.UserID).Preload("Roles").First(&dbUser).Error; err != nil {
		log.Printf("user not found: %s", err)
		return &auth.AuthenticateResponse{
			Authenticated: false,
			User:          nil,
		}, nil
	}

	var roles []string
	for _, role := range dbUser.Roles {
		roles = append(roles, role.Title)
	}

	authUser := auth.User{
		Id:        dbUser.ID.String(),
		FirstName: dbUser.FirstName,
		LastName:  dbUser.LastName,
		Phone:     dbUser.Phone,
		Email:     dbUser.Email,
		Roles:     roles,
	}

	return &auth.AuthenticateResponse{
		Authenticated: true,
		User:          &authUser,
	}, nil
}
