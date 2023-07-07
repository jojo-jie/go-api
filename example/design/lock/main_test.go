package main

import (
	"sync"
	"sync/atomic"
	"testing"
)

func TestSelfLock(t *testing.T) {
	var sl spinLock
	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			sl.Lock(t)
			defer sl.Unlock()
			t.Logf("index %d got spin lock\n", idx)
		}(i)
	}
	wg.Wait()
}

type spinLock uint32

// Lock 获取自旋锁
func (sl *spinLock) Lock(t *testing.T) {
	for !atomic.CompareAndSwapUint32((*uint32)(sl), 0, 1) {
		// 获取到自旋
		t.Log("获取到自旋")
	}
}

// Unlock 释放自旋锁
func (sl *spinLock) Unlock() {
	atomic.StoreUint32((*uint32)(sl), 0)
}
