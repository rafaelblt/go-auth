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
var dbProvider *testutil.DatabaseProvider

func TestMain(m *testing.M) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	testDB, err := testutil.NewIsolatedDB(ctx)
	if err != nil {
		log.Fatalf("could not create isolated db: %v", err)
	}
	defer testDB.Close(context.Background())

	dbProvider, err = testutil.NewDatabaseProvider(ctx)
	if err != nil {
		log.Fatalf("could not create database provider: %v", err)
	}
	defer dbProvider.Close(context.Background())

	code := m.Run()

	os.Exit(code)
}
