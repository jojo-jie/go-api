package service

import (
	"bytes"
	"errors"
	"github.com/minio/minio-go/v7"
	"io"
	"largefile/configs"
	"largefile/internal/oss"
	"mime/multipart"
	"net/http"
	"sync"
	"time"
)

var once sync.Once
var c *configs.Config
var m *oss.Minio

func Init(config *configs.Config) {
	once.Do(func() {
		c = config
		m, _ = oss.New(c)
	})
}

func Buckets(writer http.ResponseWriter, request *http.Request) {
	buckets, _ := m.ListBuckets(request.Context())
	Success(writer, "ok", buckets)
}

func SignUrl(writer http.ResponseWriter, request *http.Request) {
	objectName, ok := request.URL.Query()["file_name"]
	if !ok {
		Error(writer, "file_name 非法", nil)
		return
	}
	uploadUrl, err := m.PreSignedPutObject(request.Context(), c.Minio.BucketName, objectName[0], 5*time.Minute)
	imagesUrl, err := m.GetPolicyUrl(request.Context(), c.Minio.BucketName, objectName[0], time.Hour, nil)
	if err != nil {
		Error(writer, err.Error(), nil)
		return
	}
	urlMap := map[string]string{
		"upload_url": uploadUrl.String(),
		"file_url":   imagesUrl.String(),
	}
	Success(writer, "ok", urlMap)
	return
}

func Upload(writer http.ResponseWriter, request *http.Request) {
	err := request.ParseMultipartForm(2 << 20)
	if err != nil {
		Error(writer, err.Error(), nil)
		return
	}
	file, header, err := request.FormFile("file")
	if err != nil {
		Error(writer, err.Error(), nil)
		return
	}
	defer file.Close()
	url, formData, err := m.PreSignedPostPolicy(request.Context(), c.Minio.BucketName, header.Filename, 5*time.Minute)
	if err != nil {
		Error(writer, err.Error(), nil)
		return
	}
	formBuf := new(bytes.Buffer)
	w := multipart.NewWriter(formBuf)
	for k, v := range formData {
		w.WriteField(k, v)
	}
	f, err := w.CreateFormFile("file", header.Filename)
	if err != nil {
		Error(writer, err.Error(), nil)
		return
	}
	_, err = io.Copy(f, file)
	if err != nil {
		Error(writer, err.Error(), nil)
		return
	}
	w.Close()
	transport, err := minio.DefaultTransport(false)
	httpClient := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}
	req, err := http.NewRequest(http.MethodPost, url.String(), bytes.NewReader(formBuf.Bytes()))
	req.Header.Set("Content-Type", w.FormDataContentType())
	res, err := httpClient.Do(req)
	if err != nil {
		Error(writer, err.Error(), nil)
		return
	}
	defer res.Body.Close()
	fileInfo := make(map[string]any)
	fileInfo["file_name"] = header.Filename
	fileInfo["size"] = header.Size
	fileInfo["mime_header"] = header.Header
	fileInfo["url"] = url.String()
	fileInfo["form_data"] = formData
	if res.StatusCode != http.StatusNoContent {
		Error(writer, errors.New(res.Status).Error(), fileInfo)
		return
	}
	Success(writer, "success", fileInfo)
}
