package api

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/rafaelblt/go-auth/internal/testutil"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
)

var dbProvider *testutil.DatabaseProvider

func TestMain(m *testing.M) {
	ctx := context.Background()

	dbp, err := testutil.NewDatabaseProvider(ctx)
	if err != nil {
		log.Fatalf("db provider creation failed: %v", err)
	}

	dbProvider = dbp

	code := m.Run()

	if err := dbProvider.Close(ctx); err != nil {
		log.Fatalf("db provider close failed: %v", err)
	}

	os.Exit(code)
}
