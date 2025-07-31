package main

import (
	"context"
	"errors"
	"github.com/gorilla/websocket"
	"log"
	"log/slog"
	"net/http"
	"testing"
)

// 定义 WebSocket 升级器
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// 存储所有客户端连接
var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan string) // 广播消息通道

// 处理 WebSocket 连接
func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("upgrade error: %v", err)
		return
	}
	defer ws.Close()

	// 注册新客户端
	clients[ws] = true

	for {
		var msg string
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("read error: %v", err)
			delete(clients, ws)
			break
		}
		// 将消息发送到广播通道
		broadcast <- msg
	}
}

// 广播消息给所有客户端
func handleBroadcast() {
	for {
		msg := <-broadcast
		for client, _ := range clients {
			err := client.WriteJSON(msg)
			if err != nil {
				log.Printf("write error: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}

func TestWebsocket(t *testing.T) {
	http.HandleFunc("/ws", handleConnections)
	go handleBroadcast() // 启动广播 goroutine
	log.Fatal(http.ListenAndServe(":8080", nil))
}

type multiHandler []slog.Handler

// MultiHandler 把若干 slog.Handler 聚合成一个 Handler。
func MultiHandler(handlers ...slog.Handler) slog.Handler {
	return multiHandler(handlers)
}

// 2. 实现 slog.Handler 接口

// Enabled 只要任一底层 handler 需要该级别就返回 true。
func (h multiHandler) Enabled(ctx context.Context, l slog.Level) bool {
	for _, hh := range h {
		if hh.Enabled(ctx, l) {
			return true
		}
	}
	return false
}

// Handle 依次调用底层 handler 的 Handle；任一失败都会收集错误，最终合并返回。
// 注意：为避免 handler 修改 Record，传给每个 handler 的是克隆后的副本。
func (h multiHandler) Handle(ctx context.Context, r slog.Record) error {
	var errs []error
	for _, hh := range h {
		if hh.Enabled(ctx, r.Level) { // 再次过滤，确保精确投递
			if err := hh.Handle(ctx, r.Clone()); err != nil {
				errs = append(errs, err)
			}
		}
	}
	return errors.Join(errs...) // Go 1.20+
}

// WithAttrs 为每个底层 handler 生成携带额外属性的副本，并返回一个新的 multiHandler。
func (h multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, 0, len(h))
	for _, hh := range h {
		handlers = append(handlers, hh.WithAttrs(attrs))
	}
	return multiHandler(handlers)
}

// WithGroup 为每个底层 handler 生成带有命名组的副本，并返回一个新的 multiHandler。
func (h multiHandler) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, 0, len(h))
	for _, hh := range h {
		handlers = append(handlers, hh.WithGroup(name))
	}
	return multiHandler(handlers)
}
