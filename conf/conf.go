package conf

import (
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go-api/cache"
	"go-api/model"
	"go-api/util"
	"io"
	"os"
	"time"
)

// Init 初始化配置项
func Init() {
	// 从本地读取环境变量
	godotenv.Load()

	// 设置日志级别
	util.BuildLogger(os.Getenv("LOG_LEVEL"))

	// 读取翻译文件
	if err := LoadLocales("conf/locales/zh-cn.yaml"); err != nil {
		util.Log().Panic("翻译文件加载失败", err)
	}

	// http 请求日志
	date := time.Now().Format("2006-01-02")
	logsDir := "./logs"
	ret, err := util.PathExists(logsDir)
	if err != nil {
		util.Log().Panic("是否存在logs目录发生错误", err)
	}
	if !ret {
		// logsDir 不存在创建
		if err=os.Mkdir(logsDir, os.ModePerm);err!=nil {
			util.Log().Panic("logs dir is create fail", err)
		}
	}
	path := logsDir + "/" + date + ".log"
	log, _ := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	gin.DefaultWriter = io.MultiWriter(log, os.Stdout)

	// 连接数据库
	model.Database(os.Getenv("MYSQL_DSN"))
	cache.Redis()
}
