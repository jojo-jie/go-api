package dayday

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

type value struct {
	mu    sync.Mutex
	value int
}

func TestDead(t *testing.T) {
	var wg sync.WaitGroup
	printSum := func(v1, v2 *value) {
		defer wg.Done()
		v1.mu.Lock()         //1
		defer v1.mu.Unlock() //2

		time.Sleep(2 * time.Second) //3
		v2.mu.Lock()
		defer v2.mu.Unlock()

		t.Logf("sum=%v\n", v1.value+v2.value)
	}

	var a, b value
	wg.Add(2)
	go printSum(&a, &b)
	go printSum(&b, &a)
	wg.Wait()
}

func TestPool(t *testing.T) {
	myPool := &sync.Pool{
		New: func() interface{} {
			t.Log("Creating new instance.")
			return struct{}{}
		},
	}
	myPool.Get()
	instance := myPool.Get()
	myPool.Put(instance)
	myPool.Get()
}

func TestDe(t *testing.T) {
	t.Log("return:", Demo2())
}

func Demo2() (i int) {
	defer func() {
		i++
		fmt.Println("defer2:", i) // 打印结果为 defer: 2
	}()
	defer func() {
		i++
		fmt.Println("defer1:", i) // 打印结果为 defer: 1
	}()
	return i // 或者直接 return 效果相同
}
