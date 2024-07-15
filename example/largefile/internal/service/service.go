package service

import (
	"largefile/configs"
	"largefile/internal/minio"
	"log"
	"net/http"
	"sync"
	"time"
)

var once sync.Once
var c *configs.Config
var m *minio.Minio

func Init(config *configs.Config) {
	once.Do(func() {
		log.Println("first====")
		c = config
		m, _ = minio.New(c)
	})
}

func Buckets(writer http.ResponseWriter, request *http.Request) {
	buckets, _ := m.ListBuckets(request.Context())
	writer.Write(Success("ok", buckets))
}

func PreSignUrl(writer http.ResponseWriter, request *http.Request) {
	objectName, ok := request.URL.Query()["file_name"]
	if !ok {
		writer.Write(Error("file_name 非法", nil))
		return
	}
	url, err := m.PreSignedPutObject(request.Context(), c.Minio.BucketName, objectName[0], 5*time.Minute)
	if err != nil {
		writer.Write(Error(err.Error(), nil))
		return
	}
	urlMap := map[string]string{
		"url": url.String(),
	}
	writer.Write(Success("ok", urlMap))
	return
}
