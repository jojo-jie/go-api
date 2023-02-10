package broker

import (
	"sync"
	"testing"
)

func TestTopic(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func(wg *sync.WaitGroup) {
			defer wg.Done()
			topics, err := GetTopic(NewTopicConf(WithHost("127.0.0.1:9092")))
			if err != nil {
				t.Error(err)
			}
			t.Log(topics)
		}(&wg)
	}
	wg.Wait()
}
