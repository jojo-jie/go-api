package middleware

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// GinLogger 接收gin框架默认的日志
func GinLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}

// GinRecovery recover掉项目可能出现的panic
func GinRecovery(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}
