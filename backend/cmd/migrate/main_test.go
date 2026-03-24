package main

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	migratepostgres "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestMigrationChainIncludesListingModelExpansion(t *testing.T) {
	ctx := context.Background()
	pgContainer, err := tcpostgres.Run(ctx,
		"postgres:17.8-alpine",
		tcpostgres.WithDatabase("pal_db_test"),
		tcpostgres.WithUsername("user"),
		tcpostgres.WithPassword("password"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(2*time.Minute),
		),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = pgContainer.Terminate(ctx) })

	dsn, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	driver, err := migratepostgres.WithInstance(db, &migratepostgres.Config{})
	require.NoError(t, err)

	migrator, err := migrate.NewWithDatabaseInstance(
		"file://../../db/migrations",
		"postgres",
		driver,
	)
	require.NoError(t, err)
	t.Cleanup(func() { _, _ = migrator.Close() })

	err = migrator.Up()
	require.NoError(t, err)

	requireColumnExists(t, db, "listings", "transaction_type")
	requireColumnExists(t, db, "listings", "location_province")
	requireColumnExists(t, db, "listings", "bedroom_count")
	requireColumnExists(t, db, "listings", "facilities")
	requireCheckContains(t, db, "listings", "chk_listings_status", "draft")
	requireCheckContains(t, db, "listings", "chk_listings_transaction_type", "sale")
}

func requireColumnExists(t *testing.T, db *sql.DB, table, column string) {
	t.Helper()
	var exists bool
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_name = $1 AND column_name = $2
		)
	`, table, column).Scan(&exists)
	require.NoError(t, err)
	require.True(t, exists, "expected column %s.%s to exist", table, column)
}

func requireCheckContains(t *testing.T, db *sql.DB, table, constraint, expected string) {
	t.Helper()
	var definition string
	err := db.QueryRow(`
		SELECT pg_get_constraintdef(c.oid)
		FROM pg_constraint c
		JOIN pg_class t ON c.conrelid = t.oid
		WHERE t.relname = $1 AND c.conname = $2
	`, table, constraint).Scan(&definition)
	require.NoError(t, err)
	require.Contains(t, definition, expected)
}
