package middleware

import (
	"context"
	"github.com/gin-gonic/gin"
	"go-api/serializer"
	"golang.org/x/time/rate"
	"os"
	"strconv"
	"sync"
	"time"
)

var r, b int

// 根据请求ip 限流
func Rate() gin.HandlerFunc {
	r, _ = strconv.Atoi(os.Getenv("RATE_R"))
	b, _ = strconv.Atoi(os.Getenv("RATE_B"))
	limiters:=&sync.Map{}
	return func(c *gin.Context) {
		l,_:=limiters.LoadOrStore(c.ClientIP(), rate.NewLimiter(rate.Limit(r), b))
		ctx, cancel := context.WithTimeout(c, 400*time.Millisecond)
		defer cancel()
		if err := l.(*rate.Limiter).WaitN(ctx, 1); err != nil {
			c.JSON(400, serializer.Err(serializer.CodeOverClock, "请求频率过高", err))
			c.Abort()
			return
		}
		c.Next()
	}
}
