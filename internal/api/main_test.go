package api

import (
	"context"
	"log"
	"net/http/httptest"
	"os"
	"testing"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/rafaelblt/go-auth/internal/testutil"
)

var testServer *httptest.Server
var testDB *testutil.Database

func TestMain(m *testing.M) {
	ctx := context.Background()

	db, err := testutil.NewDatabase(ctx)
	if err != nil {
		log.Fatalf("db creation failed: %v", err)
	}

	router, err := NewRouter(ctx, db.ConnectionString())
	if err != nil {
		log.Fatalf("router creation failed: %v", err)
	}

	testServer = httptest.NewServer(router)
	testDB = db

	code := m.Run()

	testServer.Close()
	testDB.Close(ctx)

	os.Exit(code)
}
