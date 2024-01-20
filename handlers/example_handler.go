package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jbrit/gojibs/models"
	"gorm.io/gorm"
)

func requireAuth(c *gin.Context, db *gorm.DB) (*models.User, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(c.GetHeader("Authorization"), claims, func(token *jwt.Token) (any, error) {
		return jwtKey, nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return nil, fmt.Errorf("")
		}
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return nil, fmt.Errorf("")
	}
	if !token.Valid {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return nil, fmt.Errorf("")
	}

	var user models.User
	if err := db.Where("id = ?", claims.UserID).First(&user).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid authentication token"})
		return nil, fmt.Errorf("")
	}
	return &user, nil
}

func CurrentUser(c *gin.Context, db *gorm.DB) {

	user, err := requireAuth(c, db)
	if err != nil {
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": user})
}
