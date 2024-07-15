package main

import (
	"context"
	"errors"
	"largefile/configs"
	"largefile/internal/routers"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var c *configs.Config

func init() {
	c = configs.New()
	err := c.Init()
	if err != nil {
		panic(err)
	}
}

func main() {
	s := &http.Server{
		Addr:           ":" + c.Server.Port,
		Handler:        routers.New(c),
		ReadTimeout:    c.Server.ReadTimeout,
		WriteTimeout:   c.Server.WriteTimeout,
		MaxHeaderBytes: 1 << 20,
	}
	log.Println("listen port:", c.Server.Port)
	go func() {
		err := s.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
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
