package infra_test

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/rafaelblt/go-auth/internal/testutil"
)

var testDB *testutil.Database

func TestMain(m *testing.M) {
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Minute)

	db, err := testutil.NewDatabase(ctx)
	if err != nil {
		log.Fatalf("could not create test database: %w", err)
	}

	testDB = db

	code := m.Run()

	ctx, _ = context.WithTimeout(context.Background(), 10*time.Second)

	if err := testDB.Finish(ctx); err != nil {
		log.Printf("test database finish failed: %v", err)
	}

	os.Exit(code)
}
