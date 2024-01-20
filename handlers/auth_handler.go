package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jbrit/gojibs/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		panic("could not hash password")
	}
	return string(hash), err
}
func IsValidHashedPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

type RegisterUserInput struct {
	Email    string `json:"email" form:"email" binding:"required,email"`
	Password string `json:"password" form:"password" binding:"required"`
}

func RegisterUser(c *gin.Context, db *gorm.DB) {
	var input RegisterUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	u, err := uuid.NewRandom()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	passwordHash, err := HashPassword(input.Password)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// TODO: handle email verification
	user := models.User{
		ID:            u.String(),
		Email:         input.Email,
		EmailVerified: false,
		PasswordHash:  passwordHash,
	}

	if tx := db.Create(&user); tx.Error != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": tx.Error.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": user})
}

var jwtKey = []byte("my_secret_key")

type Claims struct {
	UserID string `json:"id"`
	jwt.RegisteredClaims
}

type LoginUserInput struct {
	Email    string `json:"email" form:"email" binding:"required"`
	Password string `json:"password" form:"password" binding:"required"`
}

func LoginUser(c *gin.Context, db *gorm.DB) {
	var input LoginUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := db.Where("email = ?", input.Email).First(&user).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid Email/Password"})
		return
	}

	if !IsValidHashedPassword(input.Password, user.PasswordHash) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid Email/Password"})
		return
	}

	expirationTime := time.Now().Add(24 * 60 * time.Minute)
	claims := &Claims{
		UserID: user.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user, "access_token": tokenString})
}
