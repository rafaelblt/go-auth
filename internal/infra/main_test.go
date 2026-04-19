package infra_test

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/rafaelblt/go-auth/internal/testutil"
)

var testDB *testutil.IsolatedDB

func TestMain(m *testing.M) {
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Minute)

	db, err := testutil.NewIsolatedDB(ctx)
	if err != nil {
		log.Fatalf("could not create isolated db: %v", err)
	}
	defer db.Close(context.Background())

	testDB = db

	code := m.Run()

	os.Exit(code)
}
