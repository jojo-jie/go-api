package routers

import (
	"context"
	"errors"
	"github.com/gorilla/mux"
	"largefile/configs"
	"net/http"
	"time"
)

func ConfigsCtx(c *configs.Config) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), "config", c)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func Cors() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
			if r.Method != "OPTIONS" {
				next.ServeHTTP(w, r)
			}
		})
	}
}

func Response() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Date", time.Now().UTC().Format(http.TimeFormat))
			w.Header().Set("Content-Type", "application/json")
			next.ServeHTTP(w, r)
		})
	}
}

func Timeout(timeout time.Duration) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			newDeadline := time.Now().Add(timeout)
			ctx := r.Context()
			if deadline, ok := ctx.Deadline(); !ok || newDeadline.Before(deadline) {
				ctx, cancel := context.WithDeadline(ctx, newDeadline)
				defer cancel()
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			cleanCtx := context.WithoutCancel(ctx)
			newCtx, cancel := context.WithDeadline(cleanCtx, newDeadline)
			defer cancel()
			go func() {
				<-ctx.Done() // Wait for the parent context
				if !errors.Is(ctx.Err(), context.DeadlineExceeded) {
					cancel() // Cancel newCtx
				}
			}()
			next.ServeHTTP(w, r.WithContext(newCtx))
		})
	}
}
