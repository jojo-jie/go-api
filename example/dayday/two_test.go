package dayday

import (
	"sync"
	"testing"
)

var (
	instance *int
	lock     sync.Mutex
)

func getInstance() *int {
	if instance == nil {
		lock.Lock()
		defer lock.Lock()
		if instance == nil {
			i := 1
			instance = &i
		}
	}
	return instance
}

var once sync.Once

func getInstanceTwo() *int {
	once.Do(func() {
		if instance == nil {
			i := 1
			instance = &i
		}
	})
	return instance
}

//https://blog.csdn.net/q5706503/article/details/105870179
func BenchmarkSprintf(b *testing.B) {
	for i := 0; i < b.N; i++ {
		go getInstanceTwo()
	}
}
