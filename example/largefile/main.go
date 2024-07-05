package main

import (
	"fmt"
	"largefile/configs"
)

func main() {
	c := configs.New()
	err := c.Init()
	if err != nil {
		panic(err)
	}
	fmt.Println(c.Minio.BucketName)
}
