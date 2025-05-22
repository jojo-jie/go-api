package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/time/rate"
	"io"
	"log"
	"net"
	"net/http"
	"os/signal"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
)

// https://mp.weixin.qq.com/s/GG3QbKQz3wYKFPdmJjWtuA

var (
	err1 = errors.New("Error 1st")
	err2 = errors.New("Error 2nd")
)

func TestNew124(tt *testing.T) {
	timeout := 50 * time.Millisecond
	t := time.NewTimer(timeout)
	defer TrackTime()()
	time.Sleep(100 * time.Millisecond)
	t.Reset(timeout)
	<-t.C
	err := errors.Join(err1, err2)
	tt.Log(errors.Is(err, err1))
	tt.Log(errors.Is(err, err2))

	a := []int{1, 2, 3}
	b := [3]int(a[0:3])
	tt.Log(b)
}

func TestTimeA(tt *testing.T) {
	ch := make(chan int, 10)
	go func() {
		i := 1
		for {
			i++
			ch <- i
		}
	}()

	for {
		select {
		case i := <-ch:
			tt.Logf("done:%d", i)
		case <-time.After(3 * time.Minute):
			tt.Logf("现在是：%d", time.Now().Unix())
		}
	}
}

func TrackTime() func() {
	pre := time.Now()
	return func() {
		elapsed := time.Since(pre)
		fmt.Println("elapsed:", elapsed)
	}
}

func TrackTime2(pre time.Time) time.Duration {
	elapsed := time.Since(pre)
	fmt.Println("elapsed:", elapsed)
	return elapsed
}

func Ter[T any](cond bool, a, b T) T {
	if cond {
		return a
	}
	return b
}

func IsNil(x any) bool {
	if x == nil {
		return true
	}
	return reflect.ValueOf(x).IsNil()
}

func TestChain(t *testing.T) {
	s := New()

	res, err := s.HelloWorld("jojo")
	fmt.Println(res, err) // Hello World from 煎鱼

	res, err = s.HelloWorld("edd")
	fmt.Println(res, err) // name length must be greater than 3
}

type Service interface {
	HelloWorld(name string) (string, error)
}

type service struct{}

func (s service) HelloWorld(name string) (string, error) {
	return fmt.Sprintf("Hello World from %s", name), nil
}

type validator struct {
	next Service
}

func (v validator) HelloWorld(name string) (string, error) {
	if len(name) <= 3 {
		return "", fmt.Errorf("name length must be greater than 3")
	}

	return v.next.HelloWorld(name)
}

type logger struct {
	next Service
}

func (l logger) HelloWorld(name string) (string, error) {
	res, err := l.next.HelloWorld(name)

	if err != nil {
		fmt.Println("error:", err)
		return res, err
	}

	fmt.Println("HelloWorld method executed successfuly")
	return res, err
}

func New() Service {
	return logger{
		next: validator{
			next: service{},
		},
	}
}

func TestWorkerPool(t *testing.T) {
	tasks := make(chan int, 10)
	results := make(chan int, 10)
	var wg sync.WaitGroup

	// Create 3 workers
	for i := 1; i <= 3; i++ {
		wg.Add(1)
		go worker(i, tasks, results, &wg)
	}

	// Send tasks
	for i := 1; i <= 5; i++ {
		tasks <- i
	}
	close(tasks)

	// Wait for all workers to finish
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	for r := range results {
		t.Log("Result:", r)
	}
}

func worker(id int, tasks <-chan int, results chan<- int, wg *sync.WaitGroup) {
	defer wg.Done()
	for t := range tasks {
		// Do some work, e.g., multiply by 2
		results <- t * 2
	}
}

func TestZero(t *testing.T) {
	r := bytes.NewReader([]byte("Hello, 世界"))
	r.WriteTo(&zeroCopyWriter{})
}

type zeroCopyWriter struct{}

func (w *zeroCopyWriter) Write(b []byte) (int, error) {
	fmt.Printf("%v", b)
	return len(b), nil
}

func TestReadCache(t *testing.T) {
	b := []byte("Hello, 世界")
	r := bufio.NewReader(bytes.NewReader(b))
	numBytesToRead := r.Buffered()
	t.Log(numBytesToRead)
	if numBytesToRead < 5 {
		numBytesToRead = 5
	}
	p, _ := r.Peek(numBytesToRead)
	t.Log(string(p))
	_, _ = r.Discard(numBytesToRead)
}

func TestFanIn(t *testing.T) {
	t.Log(runtime.NumGoroutine())
	alice := make(chan string)
	bob := make(chan string)
	go func() {
		for i := 0; i < 3; i++ {
			alice <- fmt.Sprintf("Alice %d", i)
		}
		close(alice)
	}()
	go func() {
		for i := 0; i < 3; i++ {
			bob <- fmt.Sprintf("Bob %d", i)
		}
		close(bob)
	}()

	c := fanIn(alice, bob)
	for msg := range c {
		t.Log(msg)
	}
	t.Log(runtime.NumGoroutine())
}

// https://dev.to/shrsv/the-multiplexing-fan-in-pattern-in-go-concurrency-1c53
func fanIn(inputs ...<-chan string) <-chan string {
	c := make(chan string)
	var wg sync.WaitGroup
	wg.Add(len(inputs))

	// Launch a goroutine per input
	for _, input := range inputs {
		go func(ch <-chan string) {
			defer wg.Done()
			for v := range ch {
				c <- v
			}
		}(input)
	}

	// Close output channel when all inputs are done
	go func() {
		wg.Wait()
		close(c)
	}()
	return c
}

// Predictable Object Sizes: When you're dealing with objects of consistent size
// 可预测的对象大小：当您处理大小一致的对象时
// High-Frequency Allocations: When you're creating and destroying many objects rapidly
// 高频分配：当您快速创建和销毁许多对象时
// Short-Lived Objects: When objects are used briefly and then discarded
// 短暂对象：当对象被短暂使用后即被丢弃时
// GC Pressure: When garbage collection is causing performance issues
// 垃圾收集压力：当垃圾收集导致性能问题时
func TestSyncPool(t *testing.T) {
	data := `{"name":"Alice","age":25}`
	r := strings.NewReader(data)
	p := Pool{}
	b := p.Get(r)
	defer p.Put(b)
	var person map[string]any
	if err := b.Decode(&person); err != nil {
		t.Error(err)
	}

	t.Logf("%+v\n", person)

	// Let's make one of our circles!
	myCircle := GeometricCircle{
		// Gotta set up the embedded Point part.
		Point: Point{
			XCoordinate: 10,
			YCoordinate: 20,
		},
		// And the Circle's own bit.
		Radius: 5.5,
	}
	// You can grab the embedded fields directly! Feels like XCoordinate is part of myCircle.
	fmt.Printf("Center X: %d, Radius: %.2f\n", myCircle.XCoordinate, myCircle.Radius)
	// Output: Center X: 10, Radius: 5.50
	// And call the method from Point directly too!
	myCircle.DisplayLocation() // Output: Coordinates: [10, 20]
	fmt.Println()              // Another space... OCD maybe? haha
	// Finally, call the method that belongs only to GeometricCircle.
	myCircle.Describe() // Output: Circle with radius 5.50 centered at Coordinates: [10, 20]
}

type Pool struct {
	pool sync.Pool
}

func (p *Pool) Get(r io.Reader) *json.Decoder {
	buf := p.pool.Get()
	if buf == nil {
		return json.NewDecoder(r)
	}
	return buf.(*json.Decoder)
}

func (p *Pool) Put(decoder *json.Decoder) {
	p.pool.Put(decoder)
}

// Just a basic 'Point' struct for coordinates. Simple, right?
type Point struct {
	XCoordinate int // Let's use clear names, helps later!
	YCoordinate int
}

// Here's a method for the Point type.
func (centerPoint Point) DisplayLocation() {
	// 'centerPoint' is just the name we give the Point inside this function.
	fmt.Printf("Coordinates: [%d, %d]", centerPoint.XCoordinate, centerPoint.YCoordinate)
}

// Now, a 'Circle' struct.
// See? It *has* a center point and a radius. Makes sense.
type GeometricCircle struct {
	Point  // This is the Go magic! Embed Point directly. No field name needed!
	Radius float64
}

// And a method just for GeometricCircle.
func (circle GeometricCircle) Describe() {
	fmt.Printf("Circle with radius %.2f centered at ", circle.Radius)
	// We can totally call the method from the embedded Point! How neat is that?
	circle.DisplayLocation()
	fmt.Println() // Just adding a space for neatness.
}

// https://dev.to/tuna99/why-memory-alignment-matters-in-go-making-your-structs-lean-and-fast-1kfk
// Best Practices for Struct Layout in Go
// ✅ Go 中结构体布局的最佳实践
// Order fields from largest to smallest alignment.
// 按照从大到小的对齐顺序排列字段。
// Group fields with the same size together.
// 将相同大小的字段分组。
// Consider memory layout when defining high-volume or performance-critical structs.
// 在定义高容量或性能关键的结构体时，考虑内存布局。
// Use go vet -fieldalignment for automatic suggestions.
// 使用 go vet -fieldalignment 进行自动建议。
type PoorlyAligned struct {
	a byte
	b int32
	c int64
}

type WellAligned struct {
	c int64
	b int32
	a byte
}

var poorlySlice = make([]PoorlyAligned, 1_000_000)
var wellSlice = make([]WellAligned, 1_000_000)

func BenchmarkPoorlyAligned(b *testing.B) {
	var sum int64
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for i := range poorlySlice {
			sum += poorlySlice[i].c
		}
	}
}

func BenchmarkWellAligned(b *testing.B) {
	var sum int64
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for i := range wellSlice {
			sum += wellSlice[i].c
		}
	}
}

type User struct {
	name string
	age  int
}

// Iterator 接口
type Iterator interface {
	hasNext() bool
	getNext() *User
}

// UserCollection 用户集合
type UserCollection struct {
	users []*User
}

func (u *UserCollection) createIterator() Iterator {
	return &UserIterator{users: u.users}
}

// UserIterator 具体迭代器
type UserIterator struct {
	index int
	users []*User
}

func (u *UserIterator) hasNext() bool {
	return u.index < len(u.users)
}

func (u *UserIterator) getNext() *User {
	if u.hasNext() {
		user := u.users[u.index]
		u.index++
		return user
	}
	return nil
}

func TestIterator(t *testing.T) {
	user1 := &User{name: "Alice", age: 30}
	user2 := &User{name: "Bob", age: 25}
	collection := &UserCollection{users: []*User{user1, user2}}

	iterator := collection.createIterator()
	for iterator.hasNext() {
		user := iterator.getNext()
		t.Logf("User: %v\n", user)
	}
}

const (
	_shutdownPeriod      = 15 * time.Second
	_shutdownHardPeriod  = 3 * time.Second
	_readinessDrainDelay = 5 * time.Second
)

var isShuttingDown atomic.Bool

func TestGracefulShutdown(t *testing.T) {
	// Setup signal context
	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Readiness endpoint
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if isShuttingDown.Load() {
			http.Error(w, "Shutting down", http.StatusServiceUnavailable)
			return
		}
		fmt.Fprintln(w, "OK")
	})

	// Sample business logic
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(2 * time.Second):
			fmt.Fprintln(w, "Hello, world!")
		case <-r.Context().Done():
			http.Error(w, "Request cancelled.", http.StatusRequestTimeout)
		}
	})

	// Ensure in-flight requests aren't cancelled immediately on SIGTERM
	ongoingCtx, stopOngoingGracefully := context.WithCancel(context.Background())
	server := &http.Server{
		Addr: ":8080",
		BaseContext: func(_ net.Listener) context.Context {
			return ongoingCtx
		},
	}

	go func() {
		log.Println("Server starting on :8080.")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	// Wait for signal
	<-rootCtx.Done()
	stop()
	isShuttingDown.Store(true)
	log.Println("Received shutdown signal, shutting down.")

	// Give time for readiness check to propagate
	time.Sleep(_readinessDrainDelay)
	log.Println("Readiness check propagated, now waiting for ongoing requests to finish.")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), _shutdownPeriod)
	defer cancel()
	err := server.Shutdown(shutdownCtx)
	stopOngoingGracefully()
	if err != nil {
		log.Println("Failed to wait for ongoing requests to finish, waiting for forced cancellation.")
		time.Sleep(_shutdownHardPeriod)
	}

	log.Println("Server shut down gracefully.")
}

func TestMux(t *testing.T) {
	mux := http.NewServeMux()
	// 1. 方法匹配 (Method Matching)
	mux.HandleFunc("GET /api/users", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "获取用户列表 (GET)")
	})
	mux.HandleFunc("POST /api/users", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "创建新用户 (POST)")
	})

	// 2. 主机匹配 (Host Matching)
	mux.HandleFunc("api.example.com/data", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "来自 api.example.com 的数据服务")
	})
	mux.HandleFunc("www.example.com/data", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "来自 www.example.com 的数据展示")
	})

	// 3. 路径通配符 (Path Wildcards)
	// 单段通配符
	mux.HandleFunc("GET /users/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		fmt.Fprintf(w, "用户信息 (GET), 用户ID: %s", id)
	})
	// 多段通配符
	mux.HandleFunc("/files/{filepath...}", func(w http.ResponseWriter, r *http.Request) {
		path := r.PathValue("filepath")
		fmt.Fprintf(w, "文件路径: %s", path)
	})

	// 4. 结束匹配符 (End Matcher) 与优先级
	// 精确匹配根路径
	mux.HandleFunc("/{$}", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "精确匹配根路径")
	})
	// 匹配 /admin 结尾
	mux.HandleFunc("/admin/{$}", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "精确匹配 /admin 路径")
	})
	// 匹配所有 /admin 开头的路径 (注意尾部斜杠，优先级低于精确匹配)
	mux.HandleFunc("/admin/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "匹配所有 /admin/ 开头的路径")
	})

	// 5. 优先级规则：更具体的模式优先
	mux.HandleFunc("/assets/images/thumbnails/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "缩略图资源")
	})
	mux.HandleFunc("/assets/images/", func(w http.ResponseWriter, r *http.Request) { // 更一般的模式
		fmt.Fprintf(w, "所有图片资源")
	})

	fmt.Println("Server is listening on :8080...")
	http.ListenAndServe(":8080", mux)
}

func TestRate(t *testing.T) {
	c := make(chan int)
	r := rate.Limit(1.0)
	l := rate.NewLimiter(r, 5)
	doSomethingWithAllow(l, 10, c)
}

func doSomethingWithAllow(l *rate.Limiter, x int, c chan int) {
	if l.Allow() {
		fmt.Printf("Allowing %d to run\n", x)
	}

	c <- x
}

func processData(inputs [][]byte) {
	buffer := make([]byte, 1024) // Pre-allocate

	for _, input := range inputs {
		n := copy(buffer, input)
		fmt.Printf("Processed %d bytes\n", n)
	}
}

var bufferPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 1024)
	},
}

func handleRequest(data []byte) {
	buf := bufferPool.Get().([]byte)
	defer bufferPool.Put(buf)
	copy(buf, data)
}

func collectData(inputs []int) []int {
	result := make([]int, 0, len(inputs))

	for _, val := range inputs {
		result = append(result, val)
	}
	return result
}

func buildMessage(parts []string) string {
	var builder strings.Builder
	builder.Grow(100)
	for _, part := range parts {
		builder.WriteString(part)
	}
	return builder.String()
}

type Data struct {
	value int
}

func processData2() Data {
	d := Data{value: 42}
	return d // Stack allocated
}
func processPointer() *Data {
	d := Data{value: 42}
	return &d // Heap allocated
}

func calculate(inputs []int) int {
	sum := 0
	for _, val := range inputs {
		sum += val
	}
	return sum
}

func TestNewHttp(t *testing.T) {
	http.HandleFunc("/", OptimizedHandler)
	http.ListenAndServe(":8080", nil)
}

func OptimizedHandler(w http.ResponseWriter, r *http.Request) {
	parts := make([]string, 0, 2)
	parts = append(parts, "Hello , ")
	parts = append(parts, r.URL.Query().Get("name"))
	w.Write([]byte(buildMessage(parts)))
}
