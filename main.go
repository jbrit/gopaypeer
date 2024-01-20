package main

import (
	"github.com/gin-gonic/gin"
	"github.com/jbrit/gojibs/handlers"
	"github.com/jbrit/gojibs/models"
	"gorm.io/gorm"
)

func main() {
	router := gin.Default()
	db := models.ConnectDB()
	dbHandler := func(handler func(*gin.Context, *gorm.DB)) func(*gin.Context) {
		return func(ctx *gin.Context) { handler(ctx, db) }
	}
	router.POST("/register", dbHandler(handlers.RegisterUser))
	router.POST("/login", dbHandler(handlers.LoginUser))
	router.GET("/me", dbHandler(handlers.CurrentUser))

	router.Run("localhost:8999")
}
