package main

import (
	"go-api/conf"
	_ "go-api/docs"
	"go-api/server"
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
	// 从配置文件读取配置
	conf.Init()

	// 装载路由
	r := server.NewRouter()
	r.Run(":3000")
}
