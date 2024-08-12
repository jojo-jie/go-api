package routers

import (
	"github.com/gorilla/mux"
	"largefile/configs"
	"largefile/internal/service"
	"net/http"
)

func New(c *configs.Config) *mux.Router {
	r := mux.NewRouter()
	routers := r.PathPrefix("/minio").Subrouter()
	routers.Use(Cors(), Response())
	service.Init(c)
	routers.HandleFunc("/buckets", service.Buckets).Methods(http.MethodGet, http.MethodOptions)
	routers.HandleFunc("/binary_url", service.BinaryUrl).Methods(http.MethodGet, http.MethodOptions)
	routers.HandleFunc("/verify", service.Verify).Methods(http.MethodPost, http.MethodOptions)
	routers.HandleFunc("/remove_object", service.RemoveObject).Methods(http.MethodDelete, http.MethodOptions)
	routers.HandleFunc("/upload", service.Upload).Methods(http.MethodPost, http.MethodOptions)
	routers.HandleFunc("/chunk_upload", service.ChunkUpload).Methods(http.MethodPost, http.MethodOptions)
	routers.HandleFunc("/merge", service.Merge).Methods(http.MethodPost, http.MethodOptions)
	return routers
}
