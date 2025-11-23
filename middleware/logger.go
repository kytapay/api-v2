package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger middleware for logging requests
func Logger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf(
			"[%s] %s %s %s %d %s \"%s\" %s\n",
			param.TimeStamp.Format(time.RFC1123),
			param.ClientIP,
			param.Method,
			param.Path,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	})
}

