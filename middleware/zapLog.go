package middleware

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

var sugarLogger *zap.SugaredLogger
// GinLogger 接收gin框架默认的日志
func GinLogger() gin.HandlerFunc {
	NewZap()
	return func(c *gin.Context) {
		var data string
		if c.Request.Method == http.MethodPost {
			body,err:=c.GetRawData()
			if err!=nil {
				sugarLogger.Error(err)
			}
			c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body))
			data = string(body)
		}
		sugarLogger.Infof("方法: %s, URL: %s, CODE: %d, body数据: %s",
			c.Request.Method, c.Request.URL, c.Writer.Status(), data)
	}
}

// GinRecovery recover掉项目可能出现的panic
func GinRecovery(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}

func NewZap()  {
	var buf strings.Builder
	buf.WriteString("logs/")
	buf.WriteString(time.Now().Format("2006-01-02"))
	buf.WriteString("_zap.log")
	zf,_:=os.Create(buf.String())
	writeSyncer:=zapcore.AddSync(zf)
	encoder:=zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	core:=zapcore.NewCore(encoder, writeSyncer, zapcore.InfoLevel)
	logger:=zap.New(core)
	logger.Sugar()
	sugarLogger = logger.Sugar()
}
