package handlers

import (
	"crypto/rand"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jbrit/gojibs/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

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

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), 10)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// TODO: handle email verification
	user := models.User{
		ID:            u.String(),
		Email:         input.Email,
		EmailVerified: false,
		PasswordHash:  string(passwordHash),
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

	if !(bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)) == nil) {
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
	tokenString, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(jwtKey)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user, "access_token": tokenString})
}

type CreateOTPInput struct {
	Email  string `json:"email" form:"email" binding:"required"`
	Reason string `json:"reason" form:"reason" binding:"required,oneof=passwordreset"`
}

func CreateOTP(c *gin.Context, db *gorm.DB) {
	// TODO: rate limit this endpoint
	var input CreateOTPInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := db.Where("email = ?", input.Email).First(&user).Error; err != nil {
		// NOTE: for security reasons
		c.AbortWithStatusJSON(http.StatusCreated, gin.H{"error": "Please Check your mail for the OTP"})
		return
	}

	// generate OTP
	numbers := [...]byte{'1', '2', '3', '4', '5', '6', '7', '8', '9', '0'}
	OTP := make([]byte, 4)
	_, err := io.ReadAtLeast(rand.Reader, OTP, 4)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	for i := 0; i < len(OTP); i++ {
		OTP[i] = numbers[int(OTP[i])%len(OTP)]
	}

	user.Otp = string(OTP)
	user.OtpExpiresAt = time.Now().Add(10 * time.Minute)
	if tx := db.Save(&user); tx.Error != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": tx.Error.Error()})
		return
	}

	switch input.Reason {
	case "passwordreset":
		message := fmt.Sprintf("An attempt was made to reset your password. \nYour OTP is %s it expires in 10 minutes (%s GMT).", user.Otp, user.OtpExpiresAt.UTC().Format(time.TimeOnly))
		user.SendMail(message)
		break
	default:
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": "Please Check your mail for the OTP"})
}

var passwordResetJwtKey = []byte("my_password_secret_key")

type PasswordResetClaims struct {
	UserID string `json:"id"`
	jwt.RegisteredClaims
}
type GetPasswordChangeTokenInput struct {
	Email string `json:"email" form:"email" binding:"required"`
	OTP   string `json:"otp" form:"otp" binding:"required"`
}

func GetPasswordChangeToken(c *gin.Context, db *gorm.DB) {
	var input GetPasswordChangeTokenInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := db.Where("email = ?", input.Email).First(&user).Error; err != nil {
		// NOTE: for security reasons
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid OTP"})
		return
	}

	if err := user.ExpireOTP(input.OTP, db); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	claims := &PasswordResetClaims{
		UserID: user.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(10 * time.Minute)),
		},
	}
	tokenString, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(passwordResetJwtKey)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"password_reset_token": tokenString})
}

type PasswordChangeInput struct {
	PasswordResetToken string `json:"password_reset_token" form:"password_reset_token" binding:"required"`
	NewPassword        string `json:"new_password" form:"new_password" binding:"required"`
}

func ChangePassword(c *gin.Context, db *gorm.DB) {
	var input PasswordChangeInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: expire token after use
	passwordResetClaims := &PasswordResetClaims{}
	if err := VerifyJwt(input.PasswordResetToken, passwordResetJwtKey, passwordResetClaims); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := db.Where("id = ?", passwordResetClaims.UserID).First(&user).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid password reset token"})
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), 10)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	user.PasswordHash = string(passwordHash)
	if tx := db.Save(&user); tx.Error != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": tx.Error.Error()})
	}

	c.JSON(http.StatusOK, gin.H{"data": "Password changed successfully"})
}
