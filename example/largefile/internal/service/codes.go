package service

import "encoding/json"

type Response struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

func Success(msg string, data any) []byte {
	respByte, _ := json.Marshal(Response{
		Code: 0,
		Msg:  msg,
		Data: data,
	})
	return respByte
}

func Error(msg string, data any) []byte {
	respByte, _ := json.Marshal(Response{
		Code: -1,
		Msg:  msg,
		Data: data,
	})
	return respByte
}
