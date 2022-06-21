package dayday

import (
	"dayday/water"
	"math/rand"
	"sort"
	"sync"
	"testing"
	"time"
)

func TestWaterFactory(t *testing.T) {
	var ch chan string
	releaseHydrogen := func() {
		ch <- "H"
	}
	releaseOxygen := func() {
		ch <- "O"
	}

	var N = 100
	ch = make(chan string, N*3)

	h2o := water.New()
	var wg sync.WaitGroup
	wg.Add(N * 3)
	//h1
	go func() {
		for i := 0; i < N; i++ {
			time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
			h2o.Hydrogen(releaseHydrogen)
			wg.Done()
		}
	}()

	//h2
	go func() {
		for i := 0; i < N; i++ {
			time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
			h2o.Hydrogen(releaseHydrogen)
			wg.Done()
		}
	}()

	//o
	go func() {
		for i := 0; i < N; i++ {
			time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
			h2o.Oxygen(releaseOxygen)
			wg.Done()
		}
	}()
	wg.Wait()

	if len(ch) != N*3 {
		t.Fatalf("expect %d atom but got %d", N*3, len(ch))
	}

	var s = make([]string, 3)
	for i := 0; i < N; i++ {
		s[0] = <-ch
		s[1] = <-ch
		s[2] = <-ch
		sort.Strings(s)
		water2 := s[0] + s[1] + s[2]
		if water2 != "HHO" {
			t.Fatalf("expect a water molecule but got %s", water2)
		}
	}
}
