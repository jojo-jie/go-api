package main

import (
	"blog/global"
	"blog/internal/model"
	"blog/internal/routers"
	"blog/pkg/logger"
	"blog/pkg/setting"
	"blog/pkg/tracer"
	"context"
	"embed"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	appName     string
	isVersion   bool
	buildTime   string
	gitCommitID string
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
	setupFlag()
}

// 吸就完事了
func main() {
	if isVersion {
		fmt.Printf("app_name: %s\n", appName)
		fmt.Printf("build_version: %s\n", buildTime)
		fmt.Printf("git_commit_id: %s\n", gitCommitID)
	}

	gin.SetMode(global.ServerSetting.RunMode)
	router := routers.NewRouter()
	s := &http.Server{
		Addr:           ":" + global.ServerSetting.HttpPort,
		Handler:        router,
		ReadTimeout:    global.ServerSetting.ReadTimeOut,
		WriteTimeout:   global.ServerSetting.WriteTimeOut,
		MaxHeaderBytes: 1 << 20,
	}
	go func() {
		err := s.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("s.ListenAndServe err: %v", err)
		}
	}()

	//等待中断信号
	quit := make(chan os.Signal)
	// 接收 syscall.SIGINT syscall.SIGTERM
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shuting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown", err)
	}
	log.Println("Server exiting")
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
	jaegerTracer, _, err := tracer.NewJaegerTracer(global.ServerSetting.ServiceName, global.AppSetting.AgentHostPort)
	if err != nil {
		return err
	}
	global.Tracer = jaegerTracer
	return nil
}

func setupFlag() {
	flag.BoolVar(&isVersion, "version", false, "编译信息")
	flag.Parse()
}
