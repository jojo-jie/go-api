package main

import (
	"blog/global"
	"blog/internal/model"
	"blog/internal/routers"
	"blog/pkg/logger"
	"blog/pkg/setting"
	"blog/pkg/tracer"
	"embed"
	"github.com/gin-gonic/gin"
	"gopkg.in/natefinch/lumberjack.v2"
	"log"
	"net/http"
	"time"
)

//go:embed configs/*
var configDirs embed.FS

func init() {
	err := setupSetting(configDirs)
	if err != nil {
		log.Fatalf("init.setupSetting err: %v", err)
	}
	err = setupDBEngine()
	if err != nil {
		log.Fatalf("init.setupDBEngine err: %v", err)
	}
	err = setupLogger()
	if err != nil {
		log.Fatalf("init.setupLogger err: %v", err)
	}
	err = setupTracer()
	if err != nil {
		log.Fatalf("init.setupTracer err: %v", err)
	}

}

// 吸就完事了
func main() {
	gin.SetMode(global.ServerSetting.RunMode)
	router := routers.NewRouter()
	s := &http.Server{
		Addr:           ":" + global.ServerSetting.HttpPort,
		Handler:        router,
		ReadTimeout:    global.ServerSetting.ReadTimeOut,
		WriteTimeout:   global.ServerSetting.WriteTimeOut,
		MaxHeaderBytes: 1 << 20,
	}
	s.ListenAndServe()
}

func setupSetting(configDirs embed.FS) error {
	set, err := setting.NewSetting(configDirs)
	if err != nil {
		return err
	}
	err = set.ReadSection("Server", &global.ServerSetting)
	err = set.ReadSection("App", &global.AppSetting)
	err = set.ReadSection("Database", &global.DatabaseSetting)
	err = set.ReadSection("JWT", &global.JWTSetting)
	err = set.ReadSection("Email", &global.EmailSetting)
	if err != nil {
		return err
	}
	global.ServerSetting.ReadTimeOut *= time.Second
	global.ServerSetting.WriteTimeOut *= time.Second
	global.JWTSetting.Expire *= time.Second
	global.AppSetting.DefaultContextTimeout *= time.Second
	return nil
}

func setupDBEngine() error {
	var err error
	global.DBEngine, err = model.NewDBEngine(global.DatabaseSetting)
	if err != nil {
		return err
	}
	return nil
}

func setupLogger() error {
	global.Logger = logger.NewLogger(&lumberjack.Logger{
		Filename:  global.AppSetting.LogSavePath + "/" + global.AppSetting.LogFileName + global.AppSetting.LogFileExt,
		MaxSize:   global.AppSetting.LogMaxSize,
		MaxAge:    global.AppSetting.LogMaxAge,
		LocalTime: global.AppSetting.LogLocalTime,
	}, "", log.LstdFlags).WithCaller(2)
	return nil
}

func setupTracer() error {
	jaegerTracer, _, err := tracer.NewJaegerTracer("blog-server", "127.0.0.1:6831")
	if err != nil {
		return err
	}
	global.Tracer = jaegerTracer
	return nil
}
