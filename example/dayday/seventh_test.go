package dayday

import (
	"context"
	"golang.org/x/time/rate"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"sort"
	"sync"
	"syscall"
	"testing"
	"time"
)

type RateLimiter interface {
	Wait(ctx context.Context) error
	Limit() rate.Limit
}

func MultiLimiter(limiters ...RateLimiter) *multiLimiter {
	byLimit := func(i, j int) bool {
		return limiters[i].Limit() < limiters[j].Limit()
	}
	sort.Slice(limiters, byLimit)
	return &multiLimiter{limiters: limiters}
}

type multiLimiter struct {
	limiters []RateLimiter
}

func (l *multiLimiter) Wait(ctx context.Context) error {
	//TODO implement me
	for _, l := range l.limiters {
		if err := l.Wait(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (l *multiLimiter) Limit() rate.Limit {
	//TODO implement me
	return l.limiters[0].Limit()
}

func NewAPIConnectionV2() *APIConnectionV2 {
	secondLimit := rate.NewLimiter(Per(2, time.Second), 1)
	minuteLimit := rate.NewLimiter(Per(10, time.Minute), 10)
	return &APIConnectionV2{
		rateLimiter:  MultiLimiter(secondLimit, minuteLimit),
		diskLimit:    MultiLimiter(rate.NewLimiter(rate.Limit(1), 1)),
		networkLimit: MultiLimiter(rate.NewLimiter(Per(3, time.Second), 3)),
	}
}

type APIConnectionV2 struct {
	networkLimit,
	diskLimit,
	rateLimiter RateLimiter
}

func (a *APIConnectionV2) ReadFile(ctx context.Context) error {
	if err := MultiLimiter(a.rateLimiter, a.diskLimit).Wait(ctx); err != nil {
		return err
	}
	// Pretend we do work here
	return nil
}

func (a *APIConnectionV2) ResolveAddress(ctx context.Context) error {
	if err := MultiLimiter(a.rateLimiter, a.networkLimit).Wait(ctx); err != nil {
		return err
	}
	// Pretend we do work here
	return nil
}

func Per(eventCount int, duration time.Duration) rate.Limit {
	return rate.Every(duration / time.Duration(eventCount))
}

type Semaphore chan struct{}

func (s Semaphore) Acquire() {
	s <- struct{}{}
}

func (s Semaphore) Release() {
	<-s
}
func randomPause(max int) {
	time.Sleep(time.Millisecond * time.Duration(rand.Intn(max)))
}

// 顾客
var (
	// 控制顾客的总数
	customerMutex    sync.Mutex
	customerMaxCount = 20
	customerCount    = 0

	// 沙发容量
	sofaSema Semaphore = make(chan struct{}, 4)
)

// 理发师
var (
	// 三位理发师
	barberSema Semaphore = make(chan struct{}, 3)
)

// 收银台
var (
	// 同时只有一位理发师和顾客结账
	paySema Semaphore = make(chan struct{}, 1)
	// 顾客拿到发票后才会离开
	receiptSema Semaphore = make(chan struct{}, 1)
)

// 理发师工作
func barber(name string) {
	for {
		// 等待一个用户
		log.Println(name + "老师尝试请求第一个顾客")
		sofaSema.Release() // 等待沙发上等待最久的一位顾客
		log.Println(name + "老师找到一位顾客，开始理发")
		randomPause(2000)
		log.Println(name + "老师理完发，等待顾客付款")
		paySema.Acquire() //等待用户缴费
		log.Println(name + "老师给付完款的顾客发票")
		receiptSema.Release() // 通知顾客发票开好
		log.Println(name + "老师服务完一位顾客")
	}
}

// 模拟顾客陆陆续续的过来
func customers() {
	for {
		randomPause(500)
		go customer()
	}
}

// 顾客
func customer() {
	customerMutex.Lock()
	if customerCount == customerMaxCount {
		defer customerMutex.Unlock()
		log.Println("没有空闲座位了，一位顾客离开了")
		return
	}
	customerCount++
	customerMutex.Unlock()
	log.Println("一位顾客开始等沙发坐下")
	sofaSema.Acquire()
	log.Println("一位顾客找到空闲沙发坐下,直到被理发师叫起理发")
	paySema.Release()
	log.Println("一位顾客已付完钱")
	receiptSema.Acquire()
	log.Println("一位顾客拿到发票，离开")
	customerMutex.Lock()
	customerCount--
	customerMutex.Unlock()
}

func TestBarberShop(t *testing.T) {
	go barber("Tony")
	go barber("Kevin")
	go barber("Allen")
	go customers()
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
}
