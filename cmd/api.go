package main

import (
	"context"
	"log"
	"net/http"

	"github.com/rafaelblt/go-auth/internal/api"
	"github.com/rafaelblt/go-auth/internal/testutil"
)

func chain(h http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
    // Aplica de trás pra frente pra manter a ordem correta
    for i := len(middlewares) - 1; i >= 0; i-- {
        h = middlewares[i](h)
    }
    return h
}

func main() {
	db, err := testutil.NewDatabase(context.Background())
	if err != nil {
		log.Fatalf("db creation failed: %v", err)
	}
	router, err := api.NewRouter(context.Background(), db.ConnectionString())
	if err != nil {
		log.Fatalf("router creation failed: %v", err)
	}
	handler := chain(router,
		api.Trace,
		api.JSONContentType,
	)
	http.ListenAndServe(":8080", handler)
}
