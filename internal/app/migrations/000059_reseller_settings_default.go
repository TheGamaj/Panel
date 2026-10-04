package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddNamedMigrationContext("000059_reseller_settings_default.go", up000059ResellerSettingsDefault, emptyDown)
}

// Migration 000058 declared admins.reseller_settings as "JSON NOT NULL" on
// MySQL. MySQL cannot attach a literal default to a JSON column on older
// servers, so the column landed as NOT NULL with no default at all — the SQLite
// definition carried DEFAULT '{}', which is why the SQLite test suite never saw
// the problem.
//
// The result was that every INSERT INTO admins which did not name the column
// failed with error 1364 ("Field 'reseller_settings' doesn't have a default
// value"), including `gamaj cli admin create` and the dashboard create-admin
// endpoint.
//
// Every read path already coalesces the column (COALESCE(reseller_settings,
// '{}')), so relaxing it to nullable is safe and restores the invariant that a
// new admin can be inserted without naming this column.
func up000059ResellerSettingsDefault(ctx context.Context, tx *sql.Tx) error {
	dialect := NormalizeDialect(activeDialect())
	if dialect == "sqlite" {
		// SQLite got TEXT NOT NULL DEFAULT '{}' and is already correct.
		return nil
	}
	exists, err := HasColumn(ctx, tx, dialect, "admins", "reseller_settings")
	if err != nil || !exists {
		return err
	}
	if _, err := tx.ExecContext(ctx, `ALTER TABLE admins MODIFY COLUMN reseller_settings JSON NULL`); err != nil {
		return err
	}
	// Rows written while the column had no default hold NULL; store the same
	// empty document the column default used to supply.
	if _, err := tx.ExecContext(ctx, `UPDATE admins SET reseller_settings = '{}' WHERE reseller_settings IS NULL`); err != nil {
		return err
	}
	return nil
}
