// Command api is the go-auth service: it loads the configuration, builds the
// logger, builds the app and runs it, and shuts it down on SIGINT or SIGTERM.
//
// See docs/architecture/startup.md.
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/rafaelblt/go-auth/internal/bootstrap"
)

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	return bootstrap.Run(ctx)
}
