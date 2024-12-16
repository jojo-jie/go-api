package routes

import (
	"github.com/gorilla/mux"
	"net/http"
	"passkeydemo/internal/service"
)

func New(t service.ThirdServer) *mux.Router {
	r := mux.NewRouter()
	r.Handle("/", http.FileServer(http.Dir("static")))
	routes := r.PathPrefix("/api").Subrouter()
	routes.Use(Response())
	service.NewThirdServer(t)
	routes.HandleFunc("/register/begin", service.HandleBeginRegistration).Methods(http.MethodPost)
	routes.HandleFunc("/register/finish", service.HandleFinishRegistration).Methods(http.MethodPost)
	routes.HandleFunc("/login/begin", service.HandleBeginLogin).Methods(http.MethodPost)
	routes.HandleFunc("/login/finish", service.HandleFinishLogin).Methods(http.MethodPost)

	authRoutes := routes.PathPrefix("").Subrouter()
	authRoutes.Use(Response(), Auth(t.Session))
	authRoutes.HandleFunc("/users", service.HandleUsers).Methods(http.MethodGet)
	return r
}
