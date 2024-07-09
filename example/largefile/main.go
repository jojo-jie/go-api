package main

import (
	"context"
	"fmt"
	"largefile/configs"
	"largefile/internal/minio"
	"time"
)

var c *configs.Config

func init() {
	c = configs.New()
	err := c.Init()
	if err != nil {
		panic(err)
	}
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	client, err := minio.New(c)
	if err != nil {
		panic(err)
	}
	buckets, err := client.ListBuckets(ctx)
	if err != nil {
		panic(err)
	}
	for _, bucket := range buckets {
		fmt.Println(bucket.Name)
	}

	exists, err := client.BucketExists(ctx, c.Minio.BucketName)
	fmt.Println(exists, err)
}
