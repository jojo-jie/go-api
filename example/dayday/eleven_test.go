package dayday

import (
	"fmt"
	"github.com/bluele/gcache"
	"github.com/lucas-clemente/quic-go/http3"
	"net/http"
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
