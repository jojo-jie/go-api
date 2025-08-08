package main

import (
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// 定义WebSocket升级器
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true }, // 允许跨域，生产环境需严格配置
}

// 客户端结构体，包含连接和发送通道
type Client struct {
	conn     *websocket.Conn
	send     chan []byte
	nickname string
}

// 聊天室管理器，存储客户端和广播通道
type ChatRoom struct {
	clients    *sync.Map // 存储客户端，key为nickname，value为Client
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
}

// 创建聊天室实例
func NewChatRoom() *ChatRoom {
	return &ChatRoom{
		clients:    &sync.Map{},
		broadcast:  make(chan []byte, 256), // 缓冲区防止阻塞
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// 聊天室主循环，处理注册、注销和广播
func (cr *ChatRoom) Run() {
	for {
		select {
		case client := <-cr.register:
			cr.clients.Store(client.nickname, client)
			cr.broadcast <- []byte(client.nickname + " joined the chat")
		case client := <-cr.unregister:
			cr.clients.Delete(client.nickname)
			cr.broadcast <- []byte(client.nickname + " left the chat")
		case message := <-cr.broadcast:
			cr.clients.Range(func(key, value interface{}) bool {
				client := value.(*Client)
				select {
				case client.send <- message:
				default: // 防止阻塞
					close(client.send)
					cr.clients.Delete(key)
				}
				return true
			})
		}
	}
}

// 处理WebSocket连接
func (cr *ChatRoom) HandleWebSocket(c *gin.Context) {
	// 升级HTTP连接为WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}

	// 获取用户昵称（简化为查询参数，生产环境可用认证）
	nickname := c.Query("nickname")
	if nickname == "" {
		nickname = "Anonymous"
	}

	// 创建客户端
	client := &Client{
		conn:     conn,
		send:     make(chan []byte, 256),
		nickname: nickname,
	}

	// 注册客户端
	cr.register <- client

	// 启动读写goroutine
	go client.write()
	go client.read(cr)
}

// 客户端写消息到WebSocket
func (c *Client) write() {
	defer c.conn.Close()
	for message := range c.send {
		err := c.conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Println("Write error:", err)
			return
		}
	}
}

// 客户端从WebSocket读取消息
func (c *Client) read(cr *ChatRoom) {
	defer func() {
		cr.unregister <- c
		c.conn.Close()
	}()
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			log.Println("Read error:", err)
			return
		}
		// 广播消息
		cr.broadcast <- []byte(c.nickname + ": " + string(message))
	}
}

func main() {
	// 初始化聊天室
	chatRoom := NewChatRoom()
	go chatRoom.Run()

	// 设置Gin路由
	r := gin.Default()
	r.GET("/ws", chatRoom.HandleWebSocket)
	r.Run(":8080")
}
