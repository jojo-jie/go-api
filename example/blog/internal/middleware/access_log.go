package middleware

import (
	"blog/global"
	"blog/pkg/logger"
	"bytes"
	"github.com/gin-gonic/gin"
	"time"
)

type AccessLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w AccessLogWriter) Write(p []byte) (int, error) {
	if n, err := w.body.Write(p); err != nil {
		return n, err
	}
	return w.ResponseWriter.Write(p)
}

func AccessLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		bodyWriter:=&AccessLogWriter{
			body: bytes.NewBufferString(""),
			ResponseWriter: c.Writer,
		}
		c.Writer = bodyWriter
		beginTime:=time.Now().Unix()
		// Next Abort 执行顺序
		c.Next()
		endTime:=time.Now().Unix()
		fields:=logger.Fields{
			"request":c.Request.PostForm.Encode(),
			"response": bodyWriter.body.String(),
		}

		s:="access log: method: %s, status_code :%d, "+"begin_time:%d, end_time:%d"
		global.Logger.WithFields(fields).Infof(s, c.Request.Method, bodyWriter.Status(), beginTime, endTime)
	}
}


