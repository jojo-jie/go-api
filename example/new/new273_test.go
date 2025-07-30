package main

import (
	"github.com/gorilla/websocket"
	"log"
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
