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
	routers.Use(Response())
	service.Init(c)
	routers.HandleFunc("/buckets", service.Buckets).Methods(http.MethodGet)
	routers.HandleFunc("/binary_url", service.BinaryUrl).Methods(http.MethodGet)
	routers.HandleFunc("/verify", service.Verify).Methods(http.MethodPost)
	routers.HandleFunc("/remove_object", service.RemoveObject).Methods(http.MethodDelete)
	routers.HandleFunc("/upload", service.Upload).Methods(http.MethodPost)
	routers.HandleFunc("/chunk_upload", service.ChunkUpload).Methods(http.MethodPost)
	routers.HandleFunc("/merge", service.Merge).Methods(http.MethodPost)
	return routers
}
