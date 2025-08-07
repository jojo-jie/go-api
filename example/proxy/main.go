package main

import (
	"encoding/json"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync/atomic"
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

type ReverseProxy struct {
	backends map[uint64]*url.URL // 后端服务URL列表
	current  uint64              // 当前轮询索引，原子操作确保并发安全
}

func NewReverseProxy(backendURLs []string) *ReverseProxy {
	urls := make(map[uint64]*url.URL, len(backendURLs))
	for i, u := range backendURLs {
		parsedURL, err := url.Parse(u)
		if err != nil {
			log.Fatalf("Invalid backend URL: %v", err)
		}
		urls[uint64(i)] = parsedURL
	}
	return &ReverseProxy{backends: urls}
}

// ServeHTTP 实现http.Handler接口，选择后端并转发请求
func (p *ReverseProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 使用原子操作获取当前后端索引，实现轮询
	index := atomic.AddUint64(&p.current, 1) % uint64(len(p.backends))
	backend := p.backends[index]

	// 创建单主机反向代理
	proxy := httputil.NewSingleHostReverseProxy(backend)
	proxy.Transport = &http.Transport{
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		MaxIdleConnsPerHost:   10,
		ResponseHeaderTimeout: 10 * time.Second,
	}

	// 转发请求到选定的后端服务
	proxy.ServeHTTP(w, r)
}

func (p *ReverseProxy) healthCheck() {
	for {
		for i, backend := range p.backends {
			resp, err := http.Get(backend.String() + "/health")
			if err != nil || resp.StatusCode != http.StatusOK {
				// 标记后端不可用，动态剔除
				delete(p.backends, i)
			}
		}
		time.Sleep(10 * time.Second)
	}
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
	g.Go(func() error {
		// 定义后端服务列表
		backends := []string{
			"http://backend1:8081",
			"http://backend2:8082",
		}

		// 初始化反向代理
		proxy := NewReverseProxy(backends)
		if err := http.ListenAndServe(":8070", proxy); err != nil {
			return errors.Wrap(err, "reverse proxy server error")
		}
		return nil
	})
	if err := g.Wait(); err != nil {
		log.Fatal(err)
	}
}
