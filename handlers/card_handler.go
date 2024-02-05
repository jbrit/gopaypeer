package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func CreateCard(c *gin.Context, db *gorm.DB) {

	user, err := requireAuth(c, db)
	if err != nil {
		return
	}

	if user.DebitCard.CardNumber != "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Card Exists Already"})
		return
	}

	if err := user.CreateCard(db); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

type SetCardStausInput struct {
	CardActive bool `json:"card_active" form:"card_active"`
}

func SetCardStaus(c *gin.Context, db *gorm.DB) {
	var input SetCardStausInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	fmt.Println(input)

	user, err := requireAuth(c, db)
	if err != nil {
		return
	}

	if user.DebitCard.CardNumber == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Card does not exist"})
		return
	}
	fmt.Println(user.DebitCard)
	if user.DebitCard.CardActive == input.CardActive {
		if input.CardActive {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Card is already active"})
		} else {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Card is already blocked"})
		}
		return
	}

	user.DebitCard.CardActive = input.CardActive
	if tx := db.Save(&user.DebitCard); tx.Error != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": tx.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}
