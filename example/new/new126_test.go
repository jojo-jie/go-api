package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"new/logging"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/gorilla/mux"
)

// Config 服务器配置
type Config struct {
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// Server HTTP服务器结构体
type Server struct {
	server *http.Server
	router *mux.Router
	config Config
}

// NewServer 创建新的HTTP服务器
func NewServer(config Config) *Server {
	router := mux.NewRouter()

	s := &Server{
		router: router,
		config: config,
	}

	s.routes()

	return s
}

// routes 设置路由
func (s *Server) routes() {
	// 应用中间件
	s.router.Use(s.loggingMiddleware)
	s.router.Use(s.recoveryMiddleware)

	// 健康检查路由
	s.router.HandleFunc("/health", s.handleHealthCheck()).Methods("GET")

	// API路由组
	api := s.router.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/users", s.handleGetUsers()).Methods("GET")
	api.HandleFunc("/users/{id}", s.handleGetUser()).Methods("GET")
}

// Start 启动HTTP服务器
func (s *Server) Start() error {
	s.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", s.config.Port),
		Handler:      s.router,
		ReadTimeout:  s.config.ReadTimeout,
		WriteTimeout: s.config.WriteTimeout,
		IdleTimeout:  s.config.IdleTimeout,
	}

	// 创建错误通道
	errChan := make(chan error, 1)

	// 启动服务器
	go func() {
		log.Printf("Starting server on port %d", s.config.Port)
		if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- fmt.Errorf("server error: %w", err)
		}
	}()

	// 监听系统信号
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// 等待信号或错误
	select {
	case sig := <-signalChan:
		log.Printf("Received signal: %v", sig)
	case err := <-errChan:
		return err
	}

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	log.Println("Server gracefully stopped")
	return nil
}

// 示例处理器
func (s *Server) handleHealthCheck() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	}
}

func (s *Server) handleGetUsers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 实际业务逻辑
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"id": 1, "name": "John"}]`))
	}
}

func (s *Server) handleGetUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]
		// 实际业务逻辑
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf(`{"id": %s, "name": "User %s"}`, id, id)))
	}
}

// 中间件示例
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("Started %s %s", r.Method, r.URL.Path)

		next.ServeHTTP(w, r)

		log.Printf("Completed %s %s in %v", r.Method, r.URL.Path, time.Since(start))
	})
}

func (s *Server) recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error": "internal server error"}`))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func TestHttp(t *testing.T) {
	// 配置
	config := Config{
		Port:         8080,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	// 创建并启动服务器
	server := NewServer(config)
	if err := server.Start(); err != nil {
		t.Errorf("Server failed: %v", err)
	}
}

func TestNewErr(t *testing.T) {
	handler, err := logging.NewDailyFileHandler("./logs", "app", ".log", &slog.HandlerOptions{Level: slog.LevelDebug})
	if err != nil {
		panic(err)
	}
	defer handler.Close()

	// 设置为全局默认日志器
	slog.SetDefault(slog.New(handler))

	err = logging.Errorf("failed to process user %s", slog.String("user_id", "jojo2"), "alice")

	// 记录错误
	slog.Info("Processing failed", logging.Error(err, "jojo-group"))
}
