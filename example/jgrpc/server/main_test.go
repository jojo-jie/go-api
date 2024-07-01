package main

import (
	"context"
	"fmt"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/singleflight"
	"math/rand"
	"reflect"
	"sync/atomic"
	"testing"
	"time"
	"unsafe"
)

type MyType struct {
	Value int
}

type MyType2 struct {
	Value int
}

/*// 实现 == 操作符
func (a MyType) Equal(b MyType) bool {
	return a.Value == b.Value
}

// 实现 != 操作符
func (a MyType) NotEqual(b MyType) bool {
	return a.Value != b.Value
}*/

func TestComp(t *testing.T) {
	a := MyType{Value: 1}
	b := MyType2{Value: 1}

	t.Log(reflect.ValueOf(a).Equal(reflect.ValueOf(b)))
	v := reflect.ValueOf(a)
	t.Log(v.Type())
	t.Log(compare(a, b))
}

func compare(a, b any) bool {
	return reflect.DeepEqual(a, b)
}

func TestAto(t *testing.T) {
	var readOps uint64
	var writeOps uint64

	reads := make(chan readOp)
	writes := make(chan writeOp)

	go func() {
		var state = make(map[int]int)
		for {
			select {
			case read := <-reads:
				read.resp <- state[read.key]
			case write := <-writes:
				state[write.key] = write.val
				write.resp <- true
			}
		}
	}()

	for r := 0; r < 100; r++ {
		go func() {
			for {
				read := readOp{
					key:  rand.Intn(5),
					resp: make(chan int)}
				reads <- read
				<-read.resp
				atomic.AddUint64(&readOps, 1)
				time.Sleep(time.Millisecond)
			}
		}()
	}

	// We start 10 writes as well, using a similar
	// approach.
	for w := 0; w < 10; w++ {
		go func() {
			for {
				write := writeOp{
					key:  rand.Intn(5),
					val:  rand.Intn(100),
					resp: make(chan bool)}
				writes <- write
				<-write.resp
				atomic.AddUint64(&writeOps, 1)
				time.Sleep(time.Millisecond)
			}
		}()
	}

	time.Sleep(time.Second)

	readOpsFinal := atomic.LoadUint64(&readOps)
	fmt.Println("readOps:", readOpsFinal)
	writeOpsFinal := atomic.LoadUint64(&writeOps)
	fmt.Println("writeOps:", writeOpsFinal)
}

type readOp struct {
	key  int
	resp chan int
}
type writeOp struct {
	key  int
	val  int
	resp chan bool
}

func TestCtx(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
	printCh := make(chan int)
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(10)
	g.Go(func() error {
		doAnother(ctx, printCh, t)
		return nil
	})
	for num := range 3 {
		printCh <- num
	}
	cancel()
	err := g.Wait()
	t.Logf("doSomething: finished err:%+v\n", err)

	c1 := make(chan string, 1)
	go func() {
		time.Sleep(2 * time.Second)
		c1 <- "result 1"
	}()
	select {
	case res := <-c1:
		fmt.Println(res)
	case <-time.After(1 * time.Second):
		fmt.Println("timeout 1")
	}
	c2 := make(chan string, 1)
	go func() {
		time.Sleep(2 * time.Second)
		c2 <- "result 2"
	}()
	select {
	case res := <-c2:
		fmt.Println(res)
	case <-time.After(3 * time.Second):
		fmt.Println("timeout 2")
	}
}

func doAnother(ctx context.Context, printCh <-chan int, t *testing.T) {
	for {
		select {
		case num := <-printCh:
			t.Logf("doAnother: %d\n", num)
		case <-ctx.Done():
			if err := ctx.Err(); err != nil {
				t.Logf("doAnother err: %s\n", err)
			}
			t.Logf("doAnother: finished\n")
			return
		}
	}
}

func TestZeroStruct(t *testing.T) {
	done := make(chan struct{})
	go func() {
		t.Logf("...\n")
		close(done)
	}()
	t.Logf("waiting...\n")
	<-done
	t.Logf("end...\n")

	type Empty struct{}

	var s1 struct{}
	s2 := Empty{}
	s3 := struct{}{}

	fmt.Printf("s1 addr: %p, size: %d\n", &s1, unsafe.Sizeof(s1))
	fmt.Printf("s2 addr: %p, size: %d\n", &s2, unsafe.Sizeof(s2))
	fmt.Printf("s3 addr: %p, size: %d\n", &s3, unsafe.Sizeof(s3))
	fmt.Printf("s1 == s2 == s3: %t\n", s1 == s2 && s2 == s3)
}

func TestSingle(t *testing.T) {
	words := []string{"Go", "Go", "Go", "Rust", "PHP", "JavaScript", "Java"}
	results, err := coSearch(context.Background(), words)
	if err != nil {
		t.Error(err)
		return
	}

	t.Log(results)
}

func coSearch(ctx context.Context, words []string) ([]string, error) {
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(10)

	results := make([]string, len(words))

	for i, word := range words {
		i, word := i, word
		g.Go(func() error {
			result, err := search(ctx, word)
			if err != nil {
				return err
			}
			results[i] = result
			return nil
		})
	}

	err := g.Wait()
	return results, err
}

var g = new(singleflight.Group)

func search(ctx context.Context, word string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*3)
	defer cancel()
	result := g.DoChan(word, func() (interface{}, error) {
		return query(ctx, word)
	})
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case r := <-result:
		return r.Val.(string), r.Err
	}
}

func query(ctx context.Context, word string) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
		return fmt.Sprintf("result: %s", word), nil
	}
}
