package logger

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func GetLogger() gin.HandlerFunc {
	logger, _ := zap.NewDevelopment(zap.AddStacktrace(zap.InfoLevel))
	defer logger.Sync()
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		logger.Info("Request", zap.String("URL", c.Request.URL.String()), zap.String("method", c.Request.Method), zap.Duration("duration", time.Since(start)), zap.Int("status", c.Writer.Status()), zap.Int("size", c.Writer.Size()))
	}
}
