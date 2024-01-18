package main

import (
	"github.com/gin-gonic/gin"
	"github.com/jbrit/gojibs/handlers"
	"github.com/jbrit/gojibs/models"
)

func main() {
	router := gin.Default()
	db := models.ConnectDB()
	router.POST("/example", func(ctx *gin.Context) { handlers.CreateExample(ctx, db) })
	router.GET("/examples", func(ctx *gin.Context) { handlers.GetExamples(ctx, db) })
	router.Run("localhost:8999")
}
