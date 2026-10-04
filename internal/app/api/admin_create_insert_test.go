package api

import (
	"context"
	"database/sql"
	"path/filepath"
	"strings"
	"testing"

	"github.com/TheGamaj/Panel/internal/app/migrations"
)

// TestInsertAdminSQLNamesResellerSettings runs the dashboard's real insert
// statement — the one handleCreateAdmin executes — against a table rebuilt to
// match what MySQL actually produces for migration 000058: reseller_settings
// declared NOT NULL with no default, because MySQL cannot attach a literal
// default to a JSON column on older servers.
//
// SQLite normally gets TEXT NOT NULL DEFAULT '{}' and fills the column in on
// its own, which is why the ordinary suite never saw the defect that made
// `gamaj cli admin create` and the dashboard create-admin endpoint both fail
// with error 1364.
//
// It also pins the placeholder count to the argument count, since adding a
// literal column value can silently shift one relative to the other.
func TestInsertAdminSQLNamesResellerSettings(t *testing.T) {
	db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "strict-admins.sqlite3"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	ctx := context.Background()
	if err := migrations.RunMigrations(ctx, db, "sqlite"); err != nil {
		t.Fatal(err)
	}

	strictAdminsResellerTable(t, db)

	// Eighteen arguments, matching the placeholders in insertAdminSQL.
	args := []any{
		"asli",         // username
		"root",         // created_by
		"hash",         // hashed_password
		"full_access",  // role
		`{}`,           // permissions
		"active",       // status
		nil,            // telegram_id
		nil,            // subscription_domain
		`{}`,           // subscription_settings
		nil,            // data_limit
		"used_traffic", // traffic_limit_mode
		int64(1),       // use_service_traffic_limits
		int64(1),       // show_user_traffic
		int64(0),       // delete_user_usage_limit_enabled
		nil,            // delete_user_usage_limit
		nil,            // expire
		nil,            // users_limit
		int64(0),       // require_2fa
	}
	if got, want := strings.Count(insertAdminSQL, "?"), len(args); got != want {
		t.Fatalf("insertAdminSQL has %d placeholders but the handler passes %d arguments", got, want)
	}

	if _, err := db.ExecContext(ctx, insertAdminSQL, args...); err != nil {
		t.Fatalf("dashboard admin insert must name reseller_settings or it fails on MySQL: %v", err)
	}

	var settings string
	if err := db.QueryRow(`SELECT reseller_settings FROM admins WHERE username = ?`, "asli").Scan(&settings); err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(settings) != "{}" {
		t.Fatalf("reseller_settings = %q, want {}", settings)
	}
}

// strictAdminsResellerTable rebuilds the migrated admins table with
// reseller_settings declared NOT NULL and without a default — the shape MySQL
// actually produces for migration 000058.
func strictAdminsResellerTable(t *testing.T, db *sql.DB) {
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
