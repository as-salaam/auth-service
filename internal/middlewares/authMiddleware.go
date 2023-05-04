package middlewares

import (
	"github.com/as-salaam/auth-service/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Request.Cookie(models.AuthTokenCookieName)
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		claims := &models.Claims{}
		tokenString := cookie.Value
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return models.JWTKey, nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.Set("authClaims", claims)
		c.Next()
	}
}
