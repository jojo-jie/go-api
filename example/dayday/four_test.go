package dayday

import (
	"go.uber.org/goleak"
	"runtime"
	"syscall"
	"testing"
)

type File struct{ d int }

func openFile(path string) *File {
	d, err := syscall.Open(path, syscall.O_RDONLY, 0)
	if err != nil {
		panic(err)
	}

	p := &File{d}
	runtime.SetFinalizer(p, func(p *File) {
		syscall.Close(p.d)
	})

	return p
}

func readFile(descriptor int) string {
	doSomeAllocation()
	var buf [1000]byte
	_, err := syscall.Read(descriptor, buf[:])
	if err != nil {
		panic(err)
	}

	return string(buf[:])
}

func doSomeAllocation() {
	var a *int

	// memory increase to force the GC
	for i := 0; i < 10000000; i++ {
		i := 1
		a = &i
	}

	_ = a
}

func TestGC(t *testing.T) {
	//https://mp.weixin.qq.com/s/E3Lgl9T8iQYl65T3e0vDTA
	//快速找到 Goroutine 泄露的地方
	defer goleak.VerifyNone(t)
	//https://blog.csdn.net/qcrao/article/details/121571165
	p := openFile("one_test.go")
	content := readFile(p.d)
	//runtime.KeepAlive(p)
	println("Here is the content: " + content)
}
