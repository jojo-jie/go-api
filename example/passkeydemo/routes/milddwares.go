package routes

import (
	"context"
	"github.com/boj/redistore"
	"github.com/gorilla/mux"
	"net/http"
	"passkeydemo/internal/service"
	"time"
)

func ExServer(t service.ThirdServer) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), "t", t)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func Response() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Date", time.Now().UTC().Format(http.TimeFormat))
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			next.ServeHTTP(w, r)
		})
	}
}

func Auth(session *redistore.RediStore) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			s, err := session.Get(r, "skey")
			if err != nil {
				service.Error(w, err.Error())
				return
			}
			if _, ok := s.Values["login_user_id"]; ok {
				next.ServeHTTP(w, r)
			} else {
				service.Error(w, "no auth")
				return
			}
		})
	}
}
