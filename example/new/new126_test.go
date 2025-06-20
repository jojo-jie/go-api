package main

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/singleflight"
	"log"
	"log/slog"
	"maps"
	"net/http"
	"new/logging"
	"os"
	"os/signal"
	"sync"
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
	handler, err := logging.NewDailyFileHandler("./logs", "app", ".log", &slog.HandlerOptions{Level: slog.LevelDebug}, 0)
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

type Cache struct {
	data sync.Map
}

func (c *Cache) GetOrSet(key string, fetch func() string) string {
	if v, ok := c.data.Load(key); ok {
		return v.(string)
	}
	v := fetch()
	c.data.Store(key, v)
	return v
}

func TestSyncMap(t *testing.T) {
	cache := Cache{}
	fetch := func() string {
		time.Sleep(100 * time.Millisecond)
		return "user123"
	}
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cache.GetOrSet("user:123", fetch)
		}()
	}
	tm := TaskManager{}
	for i := 0; i < 3; i++ {
		wg.Add(1)
		id := fmt.Sprintf("task%d", i)
		go func() {
			defer wg.Done()
			tm.Set(id, "pending")
			time.Sleep(50 * time.Millisecond)
			tm.Set(id, "done")
		}()
	}
	wg.Wait()
	if v, ok := tm.Get("task1"); ok {
		t.Log(v) // "done"
	}
	t.Log(Double(2))
}

type TaskManager struct {
	tasks sync.Map
}

func (tm *TaskManager) Set(id string, status string) {
	tm.tasks.Store(id, status)
}

func (tm *TaskManager) Get(id string) (string, bool) {
	v, ok := tm.tasks.Load(id)
	return v.(string), ok
}

var g singleflight.Group
var m sync.Map

func getStuff(key string) string {
	if v, ok := m.Load(key); ok {
		return v.(string)
	}
	v, _, _ := g.Do(key, func() (interface{}, error) {
		return "data", nil
	})
	m.Store(key, v)
	return v.(string)
}

type OnlyInt interface {
	int | int8 | int16 | int32 | int64
}

func Double[T OnlyInt](v T) T {
	return v * 2
}

func TestBuy(t *testing.T) {
	acc := NewAccounts(map[string]int{
		"alice": 50,
	})
	castle := LegoSet{name: "Castle", price: 40}
	plants := LegoSet{name: "Plants", price: 20}

	var g errgroup.Group
	var mu sync.Mutex
	g.Go(func() error {
		mu.Lock()
		defer mu.Unlock()
		balance := acc.Get("alice")
		if balance < castle.price {
			return errors.New("balance is less than castle")
		}
		time.Sleep(5 * time.Millisecond)
		acc.Set("alice", balance-castle.price)
		t.Log("Alice bought the castle")
		return nil
	})
	g.Go(func() error {
		mu.Lock()
		defer mu.Unlock()
		balance := acc.Get("alice")
		if balance < plants.price {
			return errors.New("balance is less than plants")
		}
		time.Sleep(10 * time.Millisecond)
		acc.Set("alice", balance-plants.price)
		t.Log("Alice bought the plants")
		return nil
	})

	if err := g.Wait(); err != nil {
		t.Error(err)
	}

	balance := acc.Get("alice")
	t.Log("Alice's balance:", balance)
}

type LegoSet struct {
	name  string
	price int
}

type Accounts struct {
	bal map[string]int
	mu  sync.Mutex
}

func NewAccounts(bal map[string]int) *Accounts {
	return &Accounts{bal: maps.Clone(bal)}
}

// Get returns the user's balance.
func (a *Accounts) Get(name string) int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.bal[name]
}

// Set changes the user's balance.
func (a *Accounts) Set(name string, amount int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.bal[name] = amount
}
