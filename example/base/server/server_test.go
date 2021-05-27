package main

import (
	"testing"
)

type Message struct {
	id      int
	name    string
	address string
	phone   int
}

type Option func(msg *Message)

var DEFAULT_MESSAGE = Message{id: -1, name: "-1", address: "-1", phone: -1}

func WithID(id int) Option {
	return func(msg *Message) {
		msg.id = id
	}
}

func WithName(name string) Option {
	return func(msg *Message) {
		msg.name = name
	}
}

func WithAddress(addr string) Option {
	return func(msg *Message) {
		msg.address = addr
	}
}

func WithPhone(phone int) Option {
	return func(msg *Message) {
		msg.phone = phone
	}
}

func NewByOption(opts ...Option) Message {
	msg := DEFAULT_MESSAGE
	for _, o := range opts {
		o(&msg)
	}
	return msg
}

// go design mode option
// Functional Options Pattern（函数式选项模式）
func Test_option(t *testing.T) {
	// 如何理解函数式选项模式
	t.Log(NewByOption(WithID(2), WithName("message2"), WithAddress("cache2"), WithPhone(456)))
}
