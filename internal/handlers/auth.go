package handlers

import (
	"github.com/as-salaam/auth-service/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"time"
)

func (h *Handler) Login(c *gin.Context) {
	var credentials models.Credentials
	if err := c.ShouldBindJSON(&credentials); err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	var user models.User
	if result := h.DB.Where("email = ?", credentials.Email).First(&user); result.Error != nil {
		c.AbortWithStatus(http.StatusUnprocessableEntity)
		return
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(credentials.Password))
	if err != nil {
		c.AbortWithStatus(http.StatusUnprocessableEntity)
		return
	}

	expirationTime := time.Now().Add(models.AuthTokenCookieTTl * time.Minute)
	claims := &models.Claims{
		UserID:    user.ID,
		UserEmail: user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(models.JWTKey)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     models.AuthTokenCookieName,
		Value:    tokenString,
		Expires:  expirationTime,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	c.JSON(http.StatusOK, user)
}

func (h *Handler) Me(c *gin.Context) {
	claimsData, exist := c.Get("authClaims")
	if !exist {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	claims := claimsData.(*models.Claims)

	var user models.User
	if result := h.DB.Where("email = ?", claims.UserEmail).First(&user); result.Error != nil {
		http.SetCookie(c.Writer, &http.Cookie{
			Name:     models.AuthTokenCookieName,
			Expires:  time.Now(),
			HttpOnly: true,
		})
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *Handler) Refresh(c *gin.Context) {
	claimsData, exist := c.Get("authClaims")
	if !exist {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	claims := claimsData.(*models.Claims)

	if time.Until(claims.ExpiresAt.Time) > 30*time.Second {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	expirationTime := time.Now().Add(models.AuthTokenCookieTTl * time.Minute)
	claims.ExpiresAt = jwt.NewNumericDate(expirationTime)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(models.JWTKey)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     models.AuthTokenCookieName,
		Value:    tokenString,
		Expires:  expirationTime,
		HttpOnly: true,
	})
}

func (h *Handler) Logout(c *gin.Context) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     models.AuthTokenCookieName,
		Expires:  time.Now(),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully logged out",
	})
}
