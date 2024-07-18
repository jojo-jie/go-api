package service

import (
	"encoding/json"
	"net/http"
	"time"
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
	setHeader(w)
	w.Write(respByte)
}

func Error(w http.ResponseWriter, msg string, data any) {
	respByte, _ := json.Marshal(Response{
		Code: -1,
		Msg:  msg,
		Data: data,
	})
	setHeader(w)
	w.Write(respByte)
}

func setHeader(w http.ResponseWriter) {
	w.Header().Set("Date", time.Now().UTC().Format(http.TimeFormat))
	w.Header().Set("Content-Type", "application/json")
}
