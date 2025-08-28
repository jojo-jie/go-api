package util

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
)

var wg sync.WaitGroup

func TestBaseLock(t *testing.T) {

	ctx := context.Background()
	lockKey := "my_service_name_" + "lock"
	expiration := 10 * time.Second

	// lock.Acquire(ctx)
	count := 0
	// 50 个携程对 count 进行 +1 操作 100 次
	for i := 0; i < 50; i++ {
		client := redis.NewClient(&redis.Options{
			Addr: "localhost:6379",
		})
		// 1. 创建锁实例
		lock := NewRedisLock(client, lockKey, expiration)
		t.Logf("启动携程：%d\n", i+1)
		time.Sleep(100)
		wg.Add(1)
		goroutineId := i
		go func() {
			defer wg.Done()
			j := 0
			for j < 100 {
				// 2.1 尝试获取锁
				success, err := lock.Acquire(ctx)
				if err != nil {
					// 2.2 获取锁出错，抛出异常
					panic(err)
				}
				if !success {
					// 2.3 没有获取到锁，重新尝试获取锁
					continue
				}
				count++
				time.Sleep(1 * time.Millisecond)
				j++
				// 3. 释放锁
				release, err := lock.Release(ctx)
				if err != nil {
					panic(err)
				}
				if !release {
					t.Logf("协程 %d 解锁失败\n", goroutineId)
				}
				if count%100 == 0 {
					t.Logf("count = %d", count)
				}
			}
		}()
	}
	wg.Wait()
	t.Logf("count = %d", count)
}

func TestLockRenewal(t *testing.T) {
	ctx := context.Background()
	lockKey := "my_service_name_" + "lock"
	expiration := 6 * time.Second
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	// 1. 创建锁实例
	lock := NewRedisLock(client, lockKey, expiration)
	success, err := lock.LockAndRenewal(ctx)
	if err != nil {
		panic(err)
	}
	if !success {
		t.Log("加锁失败")
	}

	// 新建一个客户端尝试去获取锁
	client1 := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	ctx2 := context.Background()
	rLock := NewRedisLock(client1, lockKey, expiration)
	t.Log("新建一个客户端尝试获取锁")
	for {
		success, err := rLock.Acquire(ctx2)
		if err != nil {
			panic(err)
		} else if !success {
			t.Log("新客户端获取锁失败")
		} else {
			t.Log("获取锁成功")
		}

	}
}
