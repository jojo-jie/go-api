package main

import (
	"github.com/lesismal/nbio"
	"log"
)

func main() {
	engine := nbio.NewEngine(nbio.Config{
		Network:            "tcp",
		Addrs:              []string{":8888"},
		MaxWriteBufferSize: 6 * 1024 * 1024,
	})

	engine.OnOpen(func(c *nbio.Conn) {
		log.Println("OnOpen:", c.RemoteAddr().String())
	})

	engine.OnClose(func(c *nbio.Conn, err error) {
		log.Println("OnClose:", c.RemoteAddr().String())
	})

	engine.OnData(func(c *nbio.Conn, data []byte) {
		c.Write(append([]byte{}, data...))
	})

	err := engine.Start()
	if err != nil {
		log.Fatalf("nbio.Start failed: %v\n", err)
		return
	}

	defer engine.Stop()
	<-make(chan int)
}
