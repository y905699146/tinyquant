package main

import (
	"github.com/gin-gonic/gin"
)

func main() {

	router := gin.Default()
	gin.DisableConsoleColor()
	gin.SetMode(gin.DebugMode)

	router.POST("/api/start", Start)
	router.Run(":8000")
}
