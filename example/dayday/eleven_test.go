package dayday

import (
	"fmt"
	"github.com/bluele/gcache"
	"github.com/lucas-clemente/quic-go/http3"
	"io"
	"net/http"
	"os"
	"testing"
)

func TestCache(t *testing.T) {
	gc := gcache.New(20).
		LRU().
		Build()
	gc.Set("key", "ok")
	value, err := gc.Get("key")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Get:", value)
}

func TestHttp3(t *testing.T) {
	certFile := ""
	keyFile := ""
	handle := func(w http.ResponseWriter, req *http.Request) {
		_, _ = fmt.Fprintf(w, "hello world")
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", handle)
	err := http3.ListenAndServe(":443", certFile, keyFile, mux)
	fmt.Println(err)
}

func TestReusable(t *testing.T) {
	n, err := transfer()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	t.Logf("%d bytes transferred.\n", n)
}

func transfer() (n int64, err error) {
	if n, err := io.Copy(os.Stdout, os.Stdin); err != nil {
		return n, err
	}
	return n, err
}

func ioCopy(dst, src *os.File) error {
	buf := make([]byte, 32768)
	for {
		nr, err := src.Read(buf)
		if nr > 0 {
			_, ew := dst.Write(buf[:nr])
			if ew != nil {
				return ew
			}
		}

		if io.EOF == err {
			return nil
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// write example.
func write(dst *os.File, buf []byte) error {
	nw, ew := dst.Write(buf)

	fmt.Printf("wrote    : %d bytes\n", nw)
	fmt.Printf("write err: %v\n", ew)

	return ew
}

// read example.
func read(src *os.File) error {
	buf := make([]byte, 1024*32) // in the middle.
	// buf := make([]byte, 148157) // defies the purpose of streaming.
	// buf := make([]byte, 8) // too many chunking.

	for {
		nr, er := src.Read(buf)
		// fmt.Printf("buf      : %q\n", buf[0:nr])
		fmt.Printf("read     : %d bytes\n", nr)
		fmt.Printf("read err : %v\n", er)

		if er == io.EOF {
			return nil
		}
		if er != nil {
			return er
		}
	}
	return nil
}
