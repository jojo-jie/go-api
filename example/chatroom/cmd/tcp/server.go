package main

import (
	"bufio"
	"fmt"
	"log"
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
	users := make(map[*User]struct{})
	for {
		select {
		case user := <-enteringChannel:
			//新用户进入
			users[user] = struct{}{}
		case user := <-leavingChannel:
			delete(users, user)
			close(user.MessageChannel)
		case msg := <-messageChannel:
			for user, _ := range users {
				if user.ID == msg.OwnerID {
					continue
				}
				user.MessageChannel <- msg.Content
			}
		}
	}
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
		d := 1 * time.Minute
		timer := time.NewTicker(d)
		for {
			select {
			case <-timer.C:
				conn.Close()
			case <-userActive:
				timer.Reset(d)
			}
		}
	}()

	// 循环读取用户输入
	input := bufio.NewScanner(conn)
	for input.Scan() {
		msg.Content = strconv.Itoa(user.ID) + ":" + input.Text()
		messageChannel <- msg
		userActive <- struct{}{}
	}
	if err := input.Err(); err != nil {
		log.Fatalln("读取数据错误", err)
	}

	//用户离开
	leavingChannel <- user
	msg.Content = "user:`" + strconv.Itoa(user.ID) + "` has left"
	messageChannel <- msg
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
