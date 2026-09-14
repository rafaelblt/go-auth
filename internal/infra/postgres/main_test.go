package postgres_test

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/rafaelblt/go-auth/internal/testutil/postgrestest"
)

var poolFactory *postgrestest.PoolFactory

func TestMain(m *testing.M) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	pf, err := postgrestest.NewPoolFactory(ctx)
	if err != nil {
		log.Fatalf("could not create pool factory: %v", err)
	}

	poolFactory = pf
	defer poolFactory.Close(context.Background())

	code := m.Run()

	os.Exit(code)
}
