package routers

import (
	"context"
	"github.com/gorilla/mux"
	"largefile/configs"
	"net/http"
)

func configsCtx(c *configs.Config) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), "config", c)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
