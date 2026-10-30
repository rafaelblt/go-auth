package postgres_test

import (
	"testing"

	"github.com/rafaelblt/go-auth/internal/infra/postgres"
	"github.com/stretchr/testify/assert"
)

func TestCheckConnString_AcceptsAURLAndAKeywordValueString(t *testing.T) {
	assert.NoError(t, postgres.CheckConnString("postgres://user:password@localhost:5432/db?sslmode=disable"))
	assert.NoError(t, postgres.CheckConnString("host=localhost user=user password=password dbname=db"))
}

func TestCheckConnString_ReturnsError(t *testing.T) {
	testCases := []struct {
		desc     string
		conn     string
		expected string
	}{
		{
			desc:     "empty",
			conn:     "",
			expected: "must not be blank",
		},
		{
			desc:     "blank",
			conn:     "  \t ",
			expected: "must not be blank",
		},
		{
			desc:     "bad port",
			conn:     "postgres://alice:s3cret@db.internal:abc/app",
			expected: `must be a PostgreSQL connection string: failed to parse as URL (invalid port ":abc" after host)`,
		},
		{
			desc:     "bad sslmode",
			conn:     "postgres://alice:s3cret@db.internal/app?sslmode=bogus",
			expected: "must be a PostgreSQL connection string: failed to configure TLS (sslmode is invalid)",
		},
		{
			desc:     "neither a URL nor keyword/value",
			conn:     "db.internal",
			expected: "must be a PostgreSQL connection string: failed to parse as keyword/value (invalid keyword/value)",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			assert.EqualError(t, postgres.CheckConnString(tC.conn), tC.expected)
		})
	}
}

func TestCheckConnString_NeverQuotesTheConnectionString(t *testing.T) {
	err := postgres.CheckConnString("postgres://alice:s3cret@db.internal/app?pool_max_conns=x")

	assert.Error(t, err)
	assert.NotContains(t, err.Error(), "s3cret")
	assert.NotContains(t, err.Error(), "alice")
	assert.NotContains(t, err.Error(), "db.internal")
}
