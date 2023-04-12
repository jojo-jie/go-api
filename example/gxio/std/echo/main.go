package main

import (
	"context"
	"github.com/labstack/echo"
	"github.com/lesismal/nbio/nbhttp"
	"github.com/lesismal/nbio/nbhttp/websocket"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func onHello(c echo.Context) error {
	return c.String(http.StatusOK, "hello world")
}

func onWebsocket(c echo.Context) error {
	w := c.Response()
	r := c.Request()
	upgrader := websocket.NewUpgrader()
	upgrader.OnMessage(func(c *websocket.Conn, messageType websocket.MessageType, data []byte) {
		// echo
		c.WriteMessage(messageType, data)
	})

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}
	wsConn := conn.(*websocket.Conn)
	wsConn.OnClose(func(c *websocket.Conn, err error) {
		log.Println("OnClose:", c.RemoteAddr().String(), err)
	})
	log.Println("OnOpen:", wsConn.RemoteAddr().String())
	return nil
}

func main() {
	e := echo.New()

	e.GET("/hello", onHello)
	e.GET("/ws", onWebsocket)

	svr := nbhttp.NewServer(nbhttp.Config{
		Network: "tcp",
		Addrs:   []string{"localhost:8080"},
	}, e, nil)

	err := svr.Start()
	if err != nil {
		log.Fatalf("nbio.Start failed: %v\n", err)
	}

	log.Println("serving [labstack/echo] on [nbio]")

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	<-interrupt

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	svr.Shutdown(ctx)
}
