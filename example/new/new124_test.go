package new

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"
)

// https://mp.weixin.qq.com/s/GG3QbKQz3wYKFPdmJjWtuA

var (
	err1 = errors.New("Error 1st")
	err2 = errors.New("Error 2nd")
)

func TestNew124(tt *testing.T) {
	timeout := 50 * time.Millisecond
	t := time.NewTimer(timeout)
	defer TrackTime()()
	time.Sleep(100 * time.Millisecond)
	t.Reset(timeout)
	<-t.C
	err := errors.Join(err1, err2)
	tt.Log(errors.Is(err, err1))
	tt.Log(errors.Is(err, err2))

	a := []int{1, 2, 3}
	b := [3]int(a[0:3])
	tt.Log(b)
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

func TrackTime() func() {
	pre := time.Now()
	return func() {
		elapsed := time.Since(pre)
		fmt.Println("elapsed:", elapsed)
	}
}

func TrackTime2(pre time.Time) time.Duration {
	elapsed := time.Since(pre)
	fmt.Println("elapsed:", elapsed)
	return elapsed
}

func Ter[T any](cond bool, a, b T) T {
	if cond {
		return a
	}
	return b
}

func IsNil(x any) bool {
	if x == nil {
		return true
	}
	return reflect.ValueOf(x).IsNil()
}

func TestChain(t *testing.T) {
	s := New()

	res, err := s.HelloWorld("jojo")
	fmt.Println(res, err) // Hello World from 煎鱼

	res, err = s.HelloWorld("edd")
	fmt.Println(res, err) // name length must be greater than 3
}

type Service interface {
	HelloWorld(name string) (string, error)
}

type service struct{}

func (s service) HelloWorld(name string) (string, error) {
	return fmt.Sprintf("Hello World from %s", name), nil
}

type validator struct {
	next Service
}

func (v validator) HelloWorld(name string) (string, error) {
	if len(name) <= 3 {
		return "", fmt.Errorf("name length must be greater than 3")
	}

	return v.next.HelloWorld(name)
}

type logger struct {
	next Service
}

func (l logger) HelloWorld(name string) (string, error) {
	res, err := l.next.HelloWorld(name)

	if err != nil {
		fmt.Println("error:", err)
		return res, err
	}

	fmt.Println("HelloWorld method executed successfuly")
	return res, err
}

func New() Service {
	return logger{
		next: validator{
			next: service{},
		},
	}
}
