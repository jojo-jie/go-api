package new

import (
	"fmt"
	"testing"
	"time"
)

func TestNew124(tt *testing.T) {
	timeout := 50 * time.Millisecond
	t := time.NewTimer(timeout)
	time.Sleep(100 * time.Millisecond)
	defer TrackTime(time.Now())
	t.Reset(timeout)
	<-t.C
}

func TestTimeA(tt *testing.T) {
	ch := make(chan int, 10)
	go func() {
		i := 1
		for {
			i++
			ch <- i
		}
	}()

	for {
		select {
		case i := <-ch:
			tt.Logf("done:%d", i)
		case <-time.After(3 * time.Minute):
			tt.Logf("现在是：%d", time.Now().Unix())
		}
	}
}

func TrackTime(pre time.Time) time.Duration {
	elapsed := time.Since(pre)
	fmt.Println("elapsed:", elapsed)
	return elapsed
}
