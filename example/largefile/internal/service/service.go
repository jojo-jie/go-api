package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/encrypt"
	"io"
	"largefile/configs"
	"largefile/internal/oss"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
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

func BinaryUrl(writer http.ResponseWriter, request *http.Request) {
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

func uploadedList(fileHash string) ([]string, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	keyMarker, uploadIDMarker := "", ""
	uploadedList := make([]string, 0, 5)
	for {
		uploads, err := m.ListMultipartUploads(ctx, c.Minio.BucketName, fileHash, keyMarker, uploadIDMarker, "", 1000)
		if err != nil {
			return nil, err
		}
		for _, object := range uploads.Uploads {
			uploadedList = append(uploadedList, object.Key)
		}

		if uploads.IsTruncated {
			keyMarker = uploads.NextKeyMarker
			uploadIDMarker = uploads.NextUploadIDMarker
		} else {
			break
		}
	}
	return uploadedList, nil
}

func RemoveObject(writer http.ResponseWriter, request *http.Request) {
	objectName, ok := request.URL.Query()["file_name"]
	if !ok {
		Error(writer, "file_name 非法", nil)
		return
	}
	opts := minio.RemoveObjectOptions{}
	err := m.RemoveObject(request.Context(), c.Minio.BucketName, objectName[0], opts)
	if err != nil {
		Error(writer, err.Error(), nil)
		return
	}
	Success(writer, "ok", nil)
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
	imagesUrl, err := m.GetPolicyUrl(request.Context(), c.Minio.BucketName, header.Filename, time.Hour, nil)
	if err != nil {
		Error(writer, err.Error(), nil)
		return
	}
	Success(writer, "success", map[string]string{
		"file_url": imagesUrl.String(),
	})
}

func Verify(writer http.ResponseWriter, request *http.Request) {
	var body map[string]any
	err := json.NewDecoder(request.Body).Decode(&body)
	if err != nil {
		Error(writer, "参数非法", nil)
		return
	}
	fileName, ok1 := body["fileName"].(string)
	fileHash, ok2 := body["fileHash"].(string)
	if !ok1 || !ok2 {
		Error(writer, "参数非法", nil)
		return
	}
	ext := filepath.Ext(fileName)
	s := strings.Builder{}
	s.WriteString(fileHash)
	s.WriteString(ext)
	opts := minio.StatObjectOptions{}
	_, err = m.StatObject(request.Context(), c.Minio.BucketName, s.String(), opts)
	uploadID := ""
	uploadedList := make([]any, 0)
	if err != nil {
		opts := minio.PutObjectOptions{}
		uploadID, err = m.NewMultipartUpload(request.Context(), c.Minio.BucketName, s.String(), opts)
		if err != nil {
			Error(writer, err.Error(), nil)
			return
		}
		Success(writer, "需要上传文件", map[string]any{
			"shouldUpload": true,
			"uploadedList": uploadedList,
			"uploadID":     uploadID,
		})
		return
	}
	Success(writer, "文件已存在", map[string]any{
		"shouldUpload": false,
		"uploadedList": uploadedList,
		"uploadID":     uploadID,
	})
	return
}

func ChunkUpload(writer http.ResponseWriter, request *http.Request) {
	file, header, err := request.FormFile("chunk")
	if err != nil {
		Error(writer, err.Error(), nil)
		return
	}
	defer file.Close()
	uploadID := request.PostForm.Get("uploadID")
	chunkHash := request.FormValue("chunkHash")
	fileHash := request.FormValue("fileHash")
	fileName := request.FormValue("fileName")
	partID, _ := strconv.Atoi(request.FormValue("partID"))
	if uploadID == "" || chunkHash == "" || fileHash == "" || fileName == "" {
		Error(writer, "缺少参数", nil)
		return
	}
	ext := filepath.Ext(fileName)
	s := strings.Builder{}
	s.WriteString(fileHash)
	s.WriteString(ext)
	objectPart, err := m.PutObjectPart(request.Context(), c.Minio.BucketName, s.String(), uploadID, partID, file, header.Size, minio.PutObjectPartOptions{
		Md5Base64:    "",
		Sha256Hex:    "",
		SSE:          encrypt.NewSSE(),
		CustomHeader: nil,
		Trailer:      nil,
	})
	if err != nil {
		Error(writer, err.Error(), nil)
		return
	}

	Success(writer, "ok", map[string]any{
		"uploadID": uploadID,
		"part": minio.CompletePart{
			PartNumber: partID,
			ETag:       objectPart.ETag,
		},
	})
}

type UploadRequest struct {
	Parts    []minio.CompletePart `json:"parts"`
	UploadID string               `json:"uploadID"`
	FileHash string               `json:"fileHash"`
	FileName string               `json:"fileName"`
}

func Merge(writer http.ResponseWriter, request *http.Request) {
	var body UploadRequest
	err := json.NewDecoder(request.Body).Decode(&body)
	if err != nil {
		Error(writer, "参数非法", nil)
		return
	}
	if body.UploadID == "" || body.FileHash == "" || body.FileName == "" {
		Error(writer, "缺少参数", nil)
		return
	}
	sort.Slice(body.Parts, func(i, j int) bool {
		return body.Parts[i].PartNumber < body.Parts[j].PartNumber
	})
	ext := filepath.Ext(body.FileName)
	s := strings.Builder{}
	s.WriteString(body.FileHash)
	s.WriteString(ext)
	objectContentType := "binary/octet-stream"
	metadata := make(map[string]string)
	metadata["Content-Type"] = objectContentType
	putopts := minio.PutObjectOptions{
		UserMetadata: metadata,
	}
	uploadInfo, err := m.CompleteMultipartUploadFinish(context.Background(), c.Minio.BucketName, s.String(), body.UploadID, body.Parts, putopts)
	if err != nil {
		Error(writer, err.Error(), nil)
		return
	}
	Success(writer, "ok", map[string]any{
		"uploadInfo": uploadInfo,
	})
	return
}
