package middleware

import (
	"github.com/gin-gonic/gin"
)

// Recovery middleware for handling panics
func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		c.JSON(500, gin.H{
			"error": "Internal server error",
		})
		c.Abort()
	})
}

