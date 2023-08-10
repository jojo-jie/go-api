package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
)

func index(w http.ResponseWriter, r *http.Request) {
	slog.New(slog.NewJSONHandler(os.Stdout, nil)).InfoContext(r.Context(), "index", r.Method, r.URL.Path, r.RemoteAddr)
	m := map[string]string{"name": "zhangsan", "age": "18"}
	fmt.Printf("m1: %v, len: %d\n", m, len(m))
	clear(m)
	fmt.Printf("m1: %v, len: %d\n", m, len(m))
	fmt.Fprintf(w, "<h1>Hello World</h1>")
}

func check(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<h1>Health check</h1>")
}

func main() {
	http.HandleFunc("/", index)
	http.HandleFunc("/health_check", check)
	fmt.Println("Server starting...")
	http.ListenAndServe(":3000", nil)
}
