package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

var (
	// 使用一个互斥信号量来同步
	cond = sync.NewCond(&sync.Mutex{})
	// 分别表示左右两个方向的计数器 (默认值为 0)
	// 也就是说，两个人碰面时，为了给对方让路，会向左或向右移动
	rightCnt, leftCnt int32
)

// 信号量加锁操作
// 两个人在移动方向时必须加锁
func takeStep(name string) {
	cond.L.Lock()
	cond.Wait()
	cond.L.Unlock()
}

// 方向移动
func move(name, dir string, cnt *int32) bool {
	fmt.Printf("%s 走到了 %v\n", name, dir)
	// 当前方向计数器+1
	atomic.AddInt32(cnt, 1)
	takeStep(name)
	// 如果当前计数器只被一个人修改过
	// 说明这个人移动了方向，但是对方未移动，此时可以让对方先走，程序直接返回即可
	if atomic.LoadInt32(cnt) == 1 {
		// 因为活锁
		// 所以代码永远执行不到这里
		fmt.Printf("%s 给对方让路成功 \n", name)
		return true
	}
	takeStep(name)
	// 当前方向计数器-1
	atomic.AddInt32(cnt, -1)
	return false
}

func giveWay(name string) {
	fmt.Printf("%s 尝试给对方让路 ... \n", name)
	// 模拟三次双方互相让路
	for i := 0; i < 3; i++ {
		if move(name, "left", &leftCnt) || move(name, "right", &rightCnt) {
			return
		}
	}
	fmt.Printf("%v 无奈地说: 咱们可以停止互相给对方让路了，你先过！\n", name)
}

func main() {
	go func() {
		// 1ms 后发出通知释放锁
		for range time.Tick(1 * time.Millisecond) {
			fmt.Println("1ms 后发出通知释放锁")
			cond.Broadcast()
		}
	}()
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		giveWay("李四")
	}()
	go func() {
		defer wg.Done()
		giveWay("张三")
	}()
	wg.Wait()
}
