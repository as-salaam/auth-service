package main

import (
	"fmt"
	"github.com/as-salaam/auth-service/internal/handlers"
	"github.com/as-salaam/auth-service/internal/middlewares"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RunHTTP(port uint, db *gorm.DB) (err error) {
	h := handlers.NewHandler(db)

	router := gin.Default()

	router.POST("/register", h.Register)
	router.POST("/login", h.Login)

	router.Use(middlewares.AuthMiddleware())

	router.GET("/me", h.Me)
	router.POST("/refresh", h.Refresh)
	router.POST("/logout", h.Logout)

	return router.Run(fmt.Sprintf(":%d", port))
}
