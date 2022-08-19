package snowflake

import (
	"sync"
	"testing"
)

const size = 10000

func TestCheck(t *testing.T) {
	w := NewWorker(5, 5)
	ch := make(chan uint64, size)
	count := size
	wg := sync.WaitGroup{}
	wg.Add(count)
	for i := 0; i < size; i++ {
		go func() {
			defer wg.Done()
			id, _ := w.NextID()
			ch <- id
		}()
	}
	wg.Wait()
	m := make(map[uint64]int)
	for i := 0; i < count; i++ {
		id := <-ch
		_, ok := m[id]
		if ok {
			t.Logf("repeat id %d", id)
			return
		}
		m[id] = i
	}
	t.Log("All", len(m), "snowflake ID Get successes!")
}
