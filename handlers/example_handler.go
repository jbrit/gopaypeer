package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jbrit/gojibs/models"
	"gorm.io/gorm"
)

type ExampleInput struct {
	Name string `json:"name" form:"name" binding:"required"`
}

func CreateExample(c *gin.Context, db *gorm.DB) {
	// receive and validate json input
	var input ExampleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	u, err := uuid.NewRandom()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	example := models.Example{
		ID:   u.String(),
		Name: input.Name,
	}
	if tx := db.Create(&example); tx.Error != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": tx.Error.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": example})
}

func GetExamples(c *gin.Context, db *gorm.DB) {
	var examples []models.Example
	tx := db.Find(&examples)
	if tx.Error != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": tx.Error.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": examples})
}
