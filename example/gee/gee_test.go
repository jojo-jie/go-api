package gee

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

func TestGeeServer(t *testing.T) {
	r := New()
	r.GET("/", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "URL.Path = %q\n", req.URL.Path)
	})
	r.GET("/hello", func(w http.ResponseWriter, req *http.Request) {
		r, _ := json.Marshal(req.Header)
		w.Write(r)
	})
	r.Run(":9000")
}
