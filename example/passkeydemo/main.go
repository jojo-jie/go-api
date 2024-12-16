package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/boj/redistore"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"golang.org/x/sync/errgroup"
	"log"
	"net/http"
	"os"
	"os/signal"
	"passkeydemo/internal/service"
	"passkeydemo/routes"
	"syscall"
	"time"
)

func main() {
	dsn := "postgres://postgres:@localhost:5432/test?sslmode=disable"
	sqlDB := sql.OpenDB(pgdriver.NewConnector(
		pgdriver.WithDSN(dsn),
		pgdriver.WithDatabase("auth"),
		pgdriver.WithUser("root"),
		pgdriver.WithPassword("123456"),
	))
	db := bun.NewDB(sqlDB, pgdialect.New())

	webAuthn, err := webauthn.New(&webauthn.Config{
		RPDisplayName: "passkeydemo",
		RPID:          "localhost",
		RPOrigins:     []string{"http://localhost:8080"},
	})
	if err != nil {
		log.Fatal(err)
	}

	store, err := redistore.NewRediStore(10, "tcp", ":6379", "", []byte("dfahfaiufdaiudx1231Ksis"))
	if err != nil {
		log.Fatalf("Failed to create session store: %v", err)
	}

	defer func() {
		fmt.Println("closing database connection")
		err := db.Close()
		_ = store.Close()
		if err != nil {
			return
		}
	}()

	s := &http.Server{
		Addr: ":8080",
		Handler: routes.New(service.ThirdServer{
			DB:      db,
			Authn:   webAuthn,
			Session: store,
		}),
		ReadTimeout:    3 * time.Second,
		WriteTimeout:   5 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	g := new(errgroup.Group)
	g.Go(func() error {

		return s.ListenAndServe()
	})
	g.Go(func() error {
		quit := make(chan os.Signal)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		log.Println("shutting down server...")
		return s.Shutdown(ctx)
	})

	if err := g.Wait(); err != nil {
		log.Fatalln(err.Error())
	}
}
