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
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
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
		_, err := w.Write([]byte(`{"status": "ok"}`))
		if err != nil {
			return
		}
	}
}

func (s *Server) handleGetUsers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 实际业务逻辑
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`[{"id": 1, "name": "John"}]`))
		if err != nil {
			return
		}
	}
}

func (s *Server) handleGetUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]
		// 实际业务逻辑
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(fmt.Sprintf(`{"id": %s, "name": "User %s"}`, id, id)))
		if err != nil {
			return
		}
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
				_, err := w.Write([]byte(`{"error": "internal server error"}`))
				if err != nil {
					return
				}
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
	defer func(handler *logging.DailyFileHandler) {
		err := handler.Close()
		if err != nil {

		}
	}(handler)

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
	var p atomic.Pointer[Accounts]
	acc := NewAccounts(map[string]int{
		"alice": 50,
	})
	castle := LegoSet{name: "Castle", price: 40}
	plants := LegoSet{name: "Plants", price: 20}

	p.Store(acc)
	var g errgroup.Group
	g.Go(func() error {
		balance := p.Load().Get("alice")
		if balance < plants.price {
			return errors.New("balance is less than plants")
		}
		newAcc := NewAccounts(map[string]int{
			"alice": balance - plants.price,
		})
		if p.CompareAndSwap(acc, newAcc) {
			t.Log("Alice bought the plants")
		}
		return nil
	})

	g.Go(func() error {
		balance := p.Load().Get("alice")
		if balance < castle.price {
			return errors.New("balance is less than castle")
		}
		newAcc := NewAccounts(map[string]int{
			"alice": balance - castle.price,
		})
		if p.CompareAndSwap(acc, newAcc) {
			t.Log("Alice bought the castle")
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		t.Error(err)
	}

	t.Log("Alice's balance:", p.Load().Get("alice"))
}

type LegoSet struct {
	name  string
	price int
}

type Accounts struct {
	bal map[string]int
}

func NewAccounts(bal map[string]int) *Accounts {
	return &Accounts{bal: maps.Clone(bal)}
}

// Get returns the user's balance.
func (a *Accounts) Get(name string) int {
	return a.bal[name]
}

// Set changes the user's balance.
func (a *Accounts) Set(name string, amount int) {
	a.bal[name] = amount
}

func (a *Accounts) CompareAndSet(name string, old, new int) bool {
	if a.bal[name] != old {
		return false
	}
	a.bal[name] = new
	return true
}

type UserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

var usernameErr = errors.New("username is too short")
var emailErr = &MyError{
	s: "email is invalid",
}
var passErr = errors.New("password is too short")

func validateRequest(req UserRequest) error {
	var errs []error
	if len(req.Username) < 3 {
		errs = append(errs, usernameErr)
	}
	if !strings.Contains(req.Email, "@") {
		errs = append(errs, emailErr)
	}
	if len(req.Password) < 6 {
		errs = append(errs, passErr)
	}
	return errors.Join(errs...)
}

type MyError struct {
	s string
}

func (e *MyError) Error() string {
	return e.s
}

func TestErrsJoin(t *testing.T) {
	req := UserRequest{
		Username: "jojo",
		Email:    "ss",
		Password: "123",
	}
	if err := validateRequest(req); err != nil {
		var e5 *MyError
		if errors.As(err, &e5) {
			t.Errorf("%+v", e5.Error())
		}
	}

}

type Request struct {
	buyer string
	set   LegoSet
}

type Purchase struct {
	buyer   string
	set     LegoSet
	balance int
	succeed bool
}

func Processor(acc map[string]int) (chan<- Request, <-chan Purchase) {
	in := make(chan Request)
	out := make(chan Purchase)
	acc = maps.Clone(acc)

	go func() {
		defer close(out) // 关闭输出通道
		for req := range in {
			// Handle the purchase.
			balance := acc[req.buyer]
			pur := Purchase{buyer: req.buyer, set: req.set, balance: balance}
			if balance >= req.set.price {
				pur.balance -= req.set.price
				pur.succeed = true
				acc[req.buyer] = pur.balance
			} else {
				pur.succeed = false
			}

			// Send the result.
			out <- pur
		}
	}()

	return in, out
}

func TestBuy2(t *testing.T) {
	const buyer = "Alice"
	acc := map[string]int{buyer: 50}

	wishlist := []LegoSet{
		{name: "Castle", price: 40},
		{name: "Plants", price: 20},
	}

	reqs, purs := Processor(acc)

	// 使用缓冲通道来确保每个响应都被接收
	resps := make(chan Purchase, len(wishlist))

	// Alice buys stuff.
	var wg sync.WaitGroup
	for _, set := range wishlist {
		wg.Add(1)
		go func(set LegoSet) {
			defer wg.Done()
			reqs <- Request{buyer: buyer, set: set}
		}(set)
	}

	// 接收所有购买结果
	go func() {
		wg.Wait()
		close(reqs)
	}()

	// 收集结果
	go func() {
		for pur := range purs {
			resps <- pur
		}
		close(resps)
	}()

	// 打印结果
	for pur := range resps {
		if pur.succeed {
			t.Logf("%s bought the %s\n", pur.buyer, pur.set.name)
			t.Logf("%s's balance: %d\n", pur.buyer, pur.balance)
		} else {
			t.Logf("%s failed to buy the %s due to insufficient balance.\n", pur.buyer, pur.set.name)
		}
	}
}

func externalService(ctx context.Context) (string, error) {
	// 模拟耗时操作（5秒）
	select {
	case <-time.After(5 * time.Second):
		return "response from external service", nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

// 带 traceID 的中间件
func withTrace(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		traceID := r.Header.Get("X-Trace-ID")
		ctx := context.WithValue(r.Context(), "traceID", traceID)
		next(w, r.WithContext(ctx))
	}
}

func TestHttpWriter_Write(t *testing.T) {
	type fields struct {
		w             http.ResponseWriter
		headerWritten bool
	}
	type args struct {
		data []byte
	}
	var tests []struct {
		name    string
		fields  fields
		args    args
		want    int
		wantErr bool
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &HttpWriter{
				w:             tt.fields.w,
				headerWritten: tt.fields.headerWritten,
			}
			got, err := w.Write(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Write() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Write() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHttpWriter_WriteHeader(t *testing.T) {
	type fields struct {
		w             http.ResponseWriter
		headerWritten bool
	}
	type args struct {
		statusCode int
	}
	var tests []struct {
		name   string
		fields fields
		args   args
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &HttpWriter{
				w:             tt.fields.w,
				headerWritten: tt.fields.headerWritten,
			}
			w.WriteHeader(tt.args.statusCode)
		})
	}
}

func TestNewSseHandler(t *testing.T) {
	var tests []struct {
		name string
		want *SseHandler
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewSseHandler(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewSseHandler() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRet(t *testing.T) {
	type args struct {
		w        http.ResponseWriter
		response any
	}
	var tests []struct {
		name string
		args args
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Ret(tt.args.w, tt.args.response)
		})
	}
}

func TestSseHandler_Factorial(t *testing.T) {
	type fields struct {
		clients map[chan string]struct{}
	}
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	var tests []struct {
		name   string
		fields fields
		args   args
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &SseHandler{
				clients: tt.fields.clients,
			}
			h.Factorial(tt.args.w, tt.args.r)
		})
	}
}

func TestSseHandler_Serve(t *testing.T) {
	type fields struct {
		clients map[chan string]struct{}
	}
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	var tests []struct {
		name   string
		fields fields
		args   args
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &SseHandler{
				clients: tt.fields.clients,
			}
			h.Serve(tt.args.w, tt.args.r)
		})
	}
}

func TestSseHandler_SimulateEvents(t *testing.T) {
	type fields struct {
		clients map[chan string]struct{}
	}
	var tests []struct {
		name    string
		fields  fields
		wantErr bool
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &SseHandler{
				clients: tt.fields.clients,
			}
			if err := h.SimulateEvents(); (err != nil) != tt.wantErr {
				t.Errorf("SimulateEvents() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_clearCosts(t *testing.T) {
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	var tests []struct {
		name string
		args args
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearCosts(tt.args.w, tt.args.r)
		})
	}
}

func Test_factorial(t *testing.T) {
	type args struct {
		n int
	}
	var tests []struct {
		name string
		args args
		want int
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := factorial(tt.args.n); got != tt.want {
				t.Errorf("factorial() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_fetchProductPrice(t *testing.T) {
	type args struct {
		productID string
	}
	var tests []struct {
		name    string
		args    args
		want    float64
		wantErr bool
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := fetchProductPrice(tt.args.productID)
			if (err != nil) != tt.wantErr {
				t.Errorf("fetchProductPrice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("fetchProductPrice() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getCost(t *testing.T) {
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	var tests []struct {
		name string
		args args
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getCost(tt.args.w, tt.args.r)
		})
	}
}

func Test_getProductPriceHandler(t *testing.T) {
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	var tests []struct {
		name string
		args args
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getProductPriceHandler(tt.args.w, tt.args.r)
		})
	}
}
