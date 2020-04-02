package server

import (
	"github.com/gin-gonic/gin"
)

func SetRouter() {
	router := gin.Default()
	gin.DisableConsoleColor()
	gin.SetMode(gin.DebugMode)

	router.POST("/api/start", Start)
	router.Run(":8080")
}

// Start is function
func Start(c *gin.Context) {
	c.String(200, "XXX")
}
