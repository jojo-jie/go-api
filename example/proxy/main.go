package main

import (
	"encoding/json"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"io"
	"log"
	"net/http"
	"time"
)

// handleProxy 处理客户端的HTTP请求并转发到目标服务器
func handleProxy(w http.ResponseWriter, r *http.Request) {
	// 创建HTTP客户端
	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        100,              // 最大空闲连接数
			IdleConnTimeout:     90 * time.Second, // 空闲连接超时
			TLSHandshakeTimeout: 10 * time.Second, // TLS握手超时
			MaxIdleConnsPerHost: 10,
		},
		Timeout: 30 * time.Second, // 总体请求超时
	}
	readAll, err := io.ReadAll(r.Body)
	if err != nil {
		return
	}

	// 构造目标请求，复制客户端请求的方法和URL
	var body map[string]any
	json.Unmarshal(readAll, &body)
	log.Printf("%s %s %s %+v", r.RemoteAddr, r.Method, r.URL, body)
	req, err := http.NewRequest(r.Method, r.URL.String(), r.Body)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// 复制客户端请求头到目标请求
	for k, v := range r.Header {
		req.Header[k] = v
	}

	// 发送请求到目标服务器
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Server Error", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close() // 确保响应Body被关闭，防止资源泄漏

	// 复制目标服务器的响应头到客户端响应
	for k, v := range resp.Header {
		w.Header()[k] = v
	}

	// 设置响应状态码
	w.WriteHeader(resp.StatusCode)

	// 将目标服务器的响应Body拷贝到客户端
	io.Copy(w, resp.Body)
}

func main() {
	g := errgroup.Group{}
	g.Go(func() error {
		http.HandleFunc("/proxy", handleProxy)
		if err := http.ListenAndServe(":8080", nil); err != nil {
			return errors.Wrap(err, "proxy server error")
		}
		return nil
	})
	if err := g.Wait(); err != nil {
		log.Fatal(err)
	}
}
