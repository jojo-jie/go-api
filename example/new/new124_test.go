package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"runtime"
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
	c := fanIn(boring("Joe"), boring("Ann"))
	for m := range c {
		t.Log(m)
	}
	t.Log("You’re both boring; I’m outta here.")
	t.Log(runtime.NumGoroutine())
}

// https://dev.to/shrsv/the-multiplexing-fan-in-pattern-in-go-concurrency-1c53
func fanIn(input1, input2 <-chan string) <-chan string {
	c := make(chan string)
	go func() {
		for m := range input1 {
			c <- m
		}
	}()
	go func() {
		for m := range input2 {
			c <- m
		}
	}()
	return c
}

func boring(name string) <-chan string {
	c := make(chan string)
	i := 0
	ticker := time.NewTicker(500 * time.Millisecond)
	go func() {
		defer ticker.Stop()
		defer close(c)
		for range ticker.C {
			c <- fmt.Sprintf("%s says %d", name, i)
			i++
			if i >= 10 {
				return
			}
		}
	}()
	return c
}
