package e2e

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/rafaelblt/go-auth/internal/bootstrap"
	"github.com/rafaelblt/go-auth/internal/config"
	"github.com/rafaelblt/go-auth/internal/shared"
	"github.com/rafaelblt/go-auth/internal/testutil/migratetest"
)

var testApp *TestApp

const (
	AccessTokenTTL  = 30 * time.Minute
	RefreshTokenTTL = 24 * time.Hour
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	db, err := migratetest.NewDatabaseWithMigrations(ctx)
	if err != nil {
		log.Fatalf("new test database failed: %s", err)
	}
	defer db.Close(ctx)

	cfg, err := config.NewConfig(config.ConfigParams{
		Address:         "localhost:8080",
		DatabaseURL:     db.ConnectionString(),
		BcryptCost:      shared.Ptr(6),
		AccessTokenTTL:  shared.Ptr(AccessTokenTTL),
		RefreshTokenTTL: shared.Ptr(RefreshTokenTTL),
	})
	if err != nil {
		log.Fatalf("new config failed: %s", err)
	}

	app, err := bootstrap.NewApp(ctx, bootstrap.AppParams{Config: cfg})
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
