package e2e

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/rafaelblt/go-auth/internal/bootstrap"
	"github.com/rafaelblt/go-auth/internal/testutil"
)

var testApp *TestApp

func TestMain(m *testing.M) {
	ctx := context.Background()

	db, err := testutil.NewDatabase(ctx)
	if err != nil {
		log.Fatalf("new test database failed: %s", err)
	}
	defer db.Close(ctx)

	cfg := bootstrap.Config{
		Address:         "localhost:8080",
		DatabaseURL:     db.ConnectionString(),
		BcryptCost:      6,
		AccessTokenTTL:  time.Minute * 30,
		RefreshTokenTTL: time.Hour * 24,
	}
	app, err := bootstrap.NewApp(ctx, cfg)
	if err != nil {
		log.Fatalf("new app failed: %s", err)
	}
	defer app.Close()

	go func() {
		if err := app.Run(ctx); err != nil {
			log.Fatalf("app run failed: %s", err)
		}
	}()

	testApp = NewTestApp(ctx, cfg)

	code := m.Run()

	os.Exit(code)
}
