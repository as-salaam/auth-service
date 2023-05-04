package handlers

import (
	"github.com/as-salaam/auth-service/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"time"
)

type UserRegistrationData struct {
	FirstName    string `json:"first_name" binding:"required"`
	LastName     string `json:"last_name" binding:"required"`
	Phone        string `json:"phone"`
	Email        string `json:"email" binding:"required"`
	Password     string `json:"password" binding:"required"`
	Confirmation string `json:"confirmation" binding:"required"`
}

func (h *Handler) Register(c *gin.Context) {
	var userData UserRegistrationData
	if err := c.ShouldBindJSON(&userData); err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	var user models.User
	if result := h.DB.Where("email = ?", userData.Email).First(&user); result.Error == nil {
		c.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{
			"message": "There is a user registered with such email already",
		})
		return
	}

	if userData.Password != userData.Confirmation {
		log.Println("passwords doesn't match")
		c.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{
			"message": "Passwords doesn't match",
		})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userData.Password), 12)
	if err != nil {
		log.Print("generating password hash:", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "Internal Server Error",
		})
		return
	}

	user.FirstName = userData.FirstName
	user.LastName = userData.LastName
	user.Phone = userData.Phone
	user.Email = userData.Email
	user.Password = string(hashedPassword)

	if result := h.DB.Create(&user); result.Error != nil {
		log.Println("inserting user data to DB:", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "Internal Server Error",
		})
		return
	}

	c.JSON(http.StatusOK, user)
}

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
