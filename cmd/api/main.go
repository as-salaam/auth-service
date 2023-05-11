package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/as-salaam/auth-service/internal/auth"
	"github.com/as-salaam/auth-service/internal/database"
	"github.com/as-salaam/auth-service/internal/handlers"
	"github.com/as-salaam/auth-service/internal/middlewares"
	"github.com/as-salaam/auth-service/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"gorm.io/gorm"
	"log"
	"net"
)

func main() {
	DBHost := flag.String("dbhost", "localhost", "Enter the host of the DB server")
	DBName := flag.String("dbname", "auth_service", "Enter the name of the DB")
	DBUser := flag.String("dbuser", "postgres", "Enter the name of a DB user")
	DBPassword := flag.String("dbpassword", "postgres", "Enter the password of user")
	DBPort := flag.Uint("dbport", 5432, "Enter the port of DB")
	Timezone := flag.String("dbtimezone", "Asia/Dushanbe", "Enter your timezone to connect to the DB")
	DBSSLMode := flag.Bool("dbsslmode", false, "Turns on ssl mode while connecting to DB")
	HTTPPort := flag.Uint("httpport", 4000, "Which port to listen")
	RPCPort := flag.Uint("rpcport", 4001, "Which port to listen")
	flag.Parse()

	db, err := database.DBInit(*DBHost, *DBName, *DBUser, *DBPassword, *DBPort, *Timezone, *DBSSLMode)
	if err != nil {
		log.Fatal("db connection:", err)
	}

	go func() {
		err = RunGRPC(*RPCPort, db)
		if err != nil {
			log.Fatal(err)
		}
	}()

	h := handlers.NewHandler(db)

	router := gin.Default()

	router.POST("/register", h.Register)
	router.POST("/login", h.Login)

	router.Use(middlewares.AuthMiddleware())

	router.GET("/me", h.Me)
	router.POST("/refresh", h.Refresh)
	router.POST("/logout", h.Logout)

	log.Fatal("router running:", router.Run(fmt.Sprintf(":%d", *HTTPPort)))
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
