package main

import (
	"fmt"
	"net"
	"strconv"
	"sync/atomic"
	"time"
)

func main() {
	listen, err := net.Listen("tcp", ":2021")
	if err != nil {
		return
	}
	go broadcaster()
	for {
		conn, err := listen.Accept()
		if err != nil {
			return
		}
		go handleConn(conn)
	}
}

type User struct {
	ID             int
	Addr           string
	EnterAt        time.Time
	MessageChannel chan string
}

func (u *User) String() string {
	return u.Addr + ", UID:" + strconv.Itoa(u.ID) + ", Enter At:" +
		u.EnterAt.Format("2006-01-02 15:04:05+8000")
}

type Message struct {
	OwnerID int
	Content string
}

var (
	//新用户到来 通过该channel 进行登记
	enteringChannel = make(chan *User)
	//用户离开，通过该 channel 进行登记
	leavingChannel = make(chan *User)
	//广播专用的用户普通消息 channel 缓冲尽可能避免出现异常情况阻塞
	messageChannel = make(chan Message, 8)
)

// 用于记录聊天室用户，并进行消息广播
func broadcaster() {

}

func handleConn(conn net.Conn) {
	defer conn.Close()
	//2006-01-02 15:04:05
	location, _ := time.LoadLocation("Asia/Shanghai")
	time.Now().In(location).Format("2006-01-02 15:04:05")
	// 新用户进入，创建该用户实例
	user := &User{
		ID:             GenUserID(),
		Addr:           conn.RemoteAddr().String(),
		EnterAt:        time.Now().In(location).Local(),
		MessageChannel: make(chan string, 8),
	}

	go sendMessage(conn, user.MessageChannel)

	user.MessageChannel <- "welcome," + user.String()
	msg := Message{
		OwnerID: user.ID,
		Content: "user:`" + strconv.Itoa(user.ID) + "` has enter",
	}
	messageChannel <- msg

	enteringChannel <- user

	//控制超时用户踢出
	var userActive = make(chan struct{})
	go func() {
		
	}()
}

var id int64

func GenUserID() int {
	return int(atomic.AddInt64(&id, 1))
}

func sendMessage(conn net.Conn, ch <-chan string) {
	for msg := range ch {
		fmt.Println(conn, msg)
	}
}
