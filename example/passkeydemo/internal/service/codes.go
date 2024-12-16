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
	respByte, _ := json.Marshal(data)
	w.WriteHeader(http.StatusOK)
	w.Write(respByte)
}

func Error(w http.ResponseWriter, msg string) {
	if msg == ErrUserNotFound {
		msg = "User does not exist"
	}
	respByte, _ := json.Marshal(msg)
	w.WriteHeader(http.StatusConflict)
	w.Write(respByte)
}
