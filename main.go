package main

import (
	"github.com/gin-contrib/pprof"
	"go-api/conf"
	_ "go-api/docs"
	"go-api/server"
	_ "go.uber.org/automaxprocs"
	"runtime"
)

// @title Gin swagger
// @version 1.0
// @description 接口文档
// @contact.name kirito
// @contact.url http://localhost/swagger/index.html
// @contact.email 18624275868@163.com
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host localhost
func main() {
	//Ballast，一种精准控制 Go GC 提高性能的方法
	//https://mp.weixin.qq.com/s/OVUsHNXGz_FicwkYgdCUdQ
	ballast := make([]byte, 10*1024*1025*1024)
	// 从配置文件读取配置
	conf.Init()

	// 装载路由
	r := server.NewRouter()
	pprof.Register(r)
	r.Run(":3000")
	runtime.KeepAlive(ballast)
}
