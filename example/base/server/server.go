package main

import (
	"fmt"
	"io"
	"net/http"
)

func main() {
	// 服务注册
	AddService(&helloService{})
	AddService(&userService{})
	http.HandleFunc("/", handler)
	http.ListenAndServe(":9988", nil)
}

func handler(writer http.ResponseWriter, request *http.Request) {
	data, _ := io.ReadAll(request.Body)
	serviceName := request.Header.Get("base-service")
	methodName := request.Header.Get("base-service-method")
	filterIvk := &filterInvoker{Invoker: &httpInvoker{}, filters: []Filter{logFilter}}
	output, _ := filterIvk.Invoke(&Invocation{
		MethodName:  methodName,
		ServiceName: serviceName,
		Input:       data,
	})
	fmt.Fprintf(writer, "%s", string(output))
}
