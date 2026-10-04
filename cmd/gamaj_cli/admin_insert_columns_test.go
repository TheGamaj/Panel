package main

import (
	"bufio"
	"context"
	"database/sql"
	"path/filepath"
	"strings"
	"testing"

	"github.com/TheGamaj/Panel/internal/app/migrations"
)

// strictAdminsTable rebuilds the migrated admins table with reseller_settings
// declared NOT NULL and *without* a default — the exact shape MySQL ends up
// with when migration 000058 runs there, because MySQL will not attach a
// literal default to a JSON column on older servers.
//
// Without this the SQLite-backed suite is blind to the problem: SQLite gets
// TEXT NOT NULL DEFAULT '{}' and happily fills the column in.
func strictAdminsTable(t *testing.T, db *sql.DB) {
	t.Helper()
	ctx := context.Background()

	var ddl string
	if err := db.QueryRowContext(ctx,
		`SELECT sql FROM sqlite_master WHERE type='table' AND name='admins'`).Scan(&ddl); err != nil {
		t.Fatal(err)
	}

	const withDefault = `"reseller_settings" TEXT NOT NULL DEFAULT '{}'`
	const withoutDefault = `"reseller_settings" TEXT NOT NULL`
	if !strings.Contains(ddl, withDefault) {
		t.Fatalf("admins DDL no longer contains %q — update this test to match the current schema:\n%s", withDefault, ddl)
	}
	strictDDL := strings.Replace(ddl, withDefault, withoutDefault, 1)

	for _, stmt := range []string{
		`ALTER TABLE admins RENAME TO admins_loose`,
		strictDDL,
		`INSERT INTO admins SELECT * FROM admins_loose`,
		`DROP TABLE admins_loose`,
	} {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			t.Fatalf("rebuilding admins with a strict reseller_settings failed on %q: %v", stmt, err)
		}
	}
}

// TestAdminCreateSurvivesStrictResellerSettingsColumn is the regression test for
// `gamaj cli admin create` failing with:
//
//	Error 1364 (HY000): Field 'reseller_settings' doesn't have a default value
//
// Every INSERT INTO admins must name this column. This runs the real
// adminCreate against a table that rejects any INSERT which omits it, so
// dropping the column from the statement turns this test red immediately.
func TestAdminCreateSurvivesStrictResellerSettingsColumn(t *testing.T) {
	db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "strict-admins.sqlite3"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := migrations.RunMigrations(context.Background(), db, "sqlite"); err != nil {
		t.Fatal(err)
	}
	strictAdminsTable(t, db)

	app := &cli{db: db, dialect: "sqlite", stdin: bufio.NewReader(strings.NewReader("\n"))}
	if err := app.adminCreate([]string{"--username", "asli", "--role", "full_access", "--password", "secret1", "--json"}); err != nil {
		t.Fatalf("admin create must name reseller_settings or it fails on MySQL: %v", err)
	}

	var settings string
	if err := db.QueryRow(`SELECT reseller_settings FROM admins WHERE username = ?`, "asli").Scan(&settings); err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(settings) != "{}" {
		t.Fatalf("reseller_settings = %q, want {}", settings)
	}
}

// TestAdminCreateStoresResellerSettingsDefault checks the value the CLI writes
// is the empty JSON document the read paths expect.
func TestAdminCreateStoresResellerSettingsDefault(t *testing.T) {
	db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "admins.sqlite3"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := migrations.RunMigrations(context.Background(), db, "sqlite"); err != nil {
		t.Fatal(err)
	}

	app := &cli{db: db, dialect: "sqlite", stdin: bufio.NewReader(strings.NewReader("\n"))}
	if err := app.adminCreate([]string{"--username", "asli", "--role", "full_access", "--password", "secret1", "--json"}); err != nil {
		t.Fatal(err)
	}

	var settings string
	if err := db.QueryRow(`SELECT reseller_settings FROM admins WHERE username = ?`, "asli").Scan(&settings); err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(settings) != "{}" {
		t.Fatalf("reseller_settings = %q, want {}", settings)
	}
}