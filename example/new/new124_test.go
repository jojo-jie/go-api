package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"runtime"
	"strings"
	"sync"
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
