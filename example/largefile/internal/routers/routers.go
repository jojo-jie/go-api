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
	//routers.Use(configsCtx(c))
	service.Init(c)
	routers.HandleFunc("/buckets", service.Buckets).Methods(http.MethodGet)
	routers.HandleFunc("/binary_url", service.BinaryUrl).Methods(http.MethodGet)
	routers.HandleFunc("/remove_object", service.RemoveObject).Methods(http.MethodDelete)
	routers.HandleFunc("/upload", service.Upload).Methods(http.MethodPost)
	routers.HandleFunc("/chunk_upload", service.ChunkUpload).Methods(http.MethodPost)
	return routers
}
