package main

import (
	"context"
	"fmt"
	"golang.org/x/sync/errgroup"
	"math/rand"
	"reflect"
	"sync/atomic"
	"testing"
	"time"
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
	g := errgroup.Group{}
	g.Go(func() error {
		doAnother(ctx, printCh, t)
		return nil
	})
	for num := range 3 {
		printCh <- num
	}
	cancel()
	g.Wait()
	t.Logf("doSomething: finished\n")
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
