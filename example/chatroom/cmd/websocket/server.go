package main

import (
	"context"
	"fmt"
	"github.com/fvbock/endless"
	"golang.org/x/sync/errgroup"
	"log"
	"net/http"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
	"os"
	"time"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		duration, err := time.ParseDuration(request.FormValue("duration"))
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		time.Sleep(duration)
		fmt.Fprint(writer, "HTTP, Hello")
	})
	mux.HandleFunc("/ws", func(writer http.ResponseWriter, request *http.Request) {
		conn, err := websocket.Accept(writer, request, nil)
		if err != nil {
			log.Fatalln(err)
		}
		defer conn.Close(websocket.StatusInternalError, "内部出错了！")
		ctx, cancel := context.WithTimeout(request.Context(), time.Second*10)
		defer cancel()
		var v interface{}
		err = wsjson.Read(ctx, conn, &v)
		if err != nil {
			log.Println(err)
			return
		}
		log.Printf("接收到客户端：%v\n", v)
		err = wsjson.Write(ctx, conn, "Hello Websocket Client")
		if err != nil {
			log.Fatal(err)
			return
		}
		conn.Close(websocket.StatusNormalClosure, "")
	})

	g := errgroup.Group{}
	g.Go(func() error {
		err := endless.ListenAndServe(":2099", mux)
		if err != nil {
			return err
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		log.Fatalln(err)
	}
	log.Println("All servers stopped. Exiting.")
	os.Exit(0)
}
