package dayday

import (
	"log"
	"sync"
)

var (
	seatsLock sync.Mutex
	seats     int
	cond      = sync.NewCond(&seatsLock)
)

func barberV2() {
	for {
		// 等待一个用户
		log.Println("Tony老师尝试请求一个顾客")
		seatsLock.Lock()
		for seats == 0 {
			cond.Wait()
		}
		seats--
		seatsLock.Unlock()
		log.Println("Tony老师找到一位顾客，开始理发")
		randomPause(2000)
	}
}

func customersV2() {
	for {
		randomPause(1000)
		go customerV2()
	}
}

func customerV2() {
	seatsLock.Lock()
	defer seatsLock.Unlock()
	if seats == 3 {
		log.Println("没有空闲座位了，一位顾客离开了")
		return
	}
	seats++
	cond.Broadcast()
	log.Println("一位顾客开始坐下排队理发")
}
