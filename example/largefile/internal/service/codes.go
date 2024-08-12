package service

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

func Success(w http.ResponseWriter, msg string, data any) {
	respByte, _ := json.Marshal(Response{
		Code: 0,
		Msg:  msg,
		Data: data,
	})
	w.Write(respByte)
}

func Error(w http.ResponseWriter, msg string, data any) {
	respByte, _ := json.Marshal(Response{
		Code: -1,
		Msg:  msg,
		Data: data,
	})
	w.Write(respByte)
}
