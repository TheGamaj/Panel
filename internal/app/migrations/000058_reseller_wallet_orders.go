package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddNamedMigrationContext("000058_reseller_wallet_orders.go", up000058ResellerWalletOrders, emptyDown)
}

func up000058ResellerWalletOrders(ctx context.Context, tx *sql.Tx) error {
	dialect := activeDialect()

	// Ensure the panel_settings table exists (fresh legacy-baseline installs
	// may not have it yet) before adding the sales columns.
	if err := createTable(ctx, tx, dialect, "panel_settings", `
CREATE TABLE panel_settings (
	id INTEGER PRIMARY KEY,
	use_nobetci INTEGER NOT NULL DEFAULT 0,
	default_subscription_type VARCHAR(32) NOT NULL DEFAULT 'key',
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
)`, `
CREATE TABLE panel_settings (
	id INTEGER NOT NULL AUTO_INCREMENT,
	use_nobetci BOOLEAN NOT NULL DEFAULT 0,
	default_subscription_type VARCHAR(32) NOT NULL DEFAULT 'key',
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY (id)
)`); err != nil {
		return err
	}

	// Gamaj sales settings: Tetraminator gateway credentials live here, set
	// by sudo admins from the dashboard; they are never hard-coded.
	for _, item := range []struct{ column, sqlite, mysql string }{
		{"tetraminator_enabled", "INTEGER NOT NULL DEFAULT 0", ""},
		{"tetraminator_base_url", "VARCHAR(512) NOT NULL DEFAULT ''", ""},
		{"tetraminator_api_key", "VARCHAR(512) NOT NULL DEFAULT ''", ""},
		{"tetraminator_min_price", "BIGINT NOT NULL DEFAULT 50000", ""},
		{"tetraminator_max_price", "BIGINT NOT NULL DEFAULT 10000000", ""},
		{"tetraminator_callback_path", "VARCHAR(256) NOT NULL DEFAULT ''", ""},
	} {
		if err := addColumn(ctx, tx, dialect, "panel_settings", item.column, item.sqlite, item.mysql); err != nil {
			return err
		}
	}

	// Reseller commerce. Quotas (data + user count) are enforced by the
	// existing admins.data_limit / admins.users_limit core, so only the
	// wallet and personalization settings are stored here.
	if err := addColumn(ctx, tx, dialect, "admins", "reseller_wallet_balance", "BIGINT NOT NULL DEFAULT 0", "BIGINT NOT NULL DEFAULT 0"); err != nil {
		return err
	}
	if err := addColumn(ctx, tx, dialect, "admins", "reseller_settings", "TEXT NOT NULL DEFAULT '{}'", "JSON NOT NULL"); err != nil {
		return err
	}

	// Immutable wallet ledger (one row per balance mutation).
	if err := createTable(ctx, tx, dialect, "reseller_wallet_transactions", `
CREATE TABLE reseller_wallet_transactions (
	id INTEGER PRIMARY KEY,
	admin_id INTEGER NOT NULL,
	amount BIGINT NOT NULL,
	balance_after BIGINT NOT NULL,
	reason VARCHAR(64) NOT NULL,
	reference VARCHAR(128) NULL,
	created_by INTEGER NULL,
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
)`, `
CREATE TABLE reseller_wallet_transactions (
	id BIGINT NOT NULL PRIMARY KEY AUTO_INCREMENT,
	admin_id BIGINT NOT NULL,
	amount BIGINT NOT NULL,
	balance_after BIGINT NOT NULL,
	reason VARCHAR(64) NOT NULL,
	reference VARCHAR(128) NULL,
	created_by BIGINT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	KEY ix_reseller_wallet_tx_admin (admin_id, id)
)`); err != nil {
		return err
	}

	// Orders created by bots / resellers for a purchasable plan.
	if err := createTable(ctx, tx, dialect, "orders", `
CREATE TABLE orders (
	id VARCHAR(64) PRIMARY KEY,
	admin_id INTEGER NOT NULL,
	plan_id INTEGER NOT NULL,
	buyer_telegram_id VARCHAR(32) NULL,
	buyer_note VARCHAR(256) NULL,
	amount BIGINT NOT NULL,
	status VARCHAR(24) NOT NULL DEFAULT 'pending',
	payment_method VARCHAR(24) NULL,
	payment_ref VARCHAR(128) NULL,
	paid_at DATETIME NULL,
	expires_at DATETIME NULL,
	provisioned_user_id INTEGER NULL,
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
)`, `
CREATE TABLE orders (
	id VARCHAR(64) NOT NULL PRIMARY KEY,
	admin_id BIGINT NOT NULL,
	plan_id BIGINT NOT NULL,
	buyer_telegram_id VARCHAR(32) NULL,
	buyer_note VARCHAR(256) NULL,
	amount BIGINT NOT NULL,
	status VARCHAR(24) NOT NULL DEFAULT 'pending',
	payment_method VARCHAR(24) NULL,
	payment_ref VARCHAR(128) NULL,
	paid_at DATETIME(6) NULL,
	expires_at DATETIME(6) NULL,
	provisioned_user_id BIGINT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	KEY ix_orders_admin (admin_id, created_at),
	KEY ix_orders_buyer (buyer_telegram_id),
	KEY ix_orders_status (status)
)`); err != nil {
		return err
	}

	// Payments attached to orders (gateway or wallet settlement).
	if err := createTable(ctx, tx, dialect, "order_payments", `
CREATE TABLE order_payments (
	id INTEGER PRIMARY KEY,
	order_id VARCHAR(64) NOT NULL,
	gateway VARCHAR(32) NOT NULL DEFAULT 'tetraminator',
	pay_id VARCHAR(128) NOT NULL,
	amount BIGINT NOT NULL,
	status VARCHAR(24) NOT NULL DEFAULT 'pending',
	inquiry_status VARCHAR(32) NULL,
	callback_seen INTEGER NOT NULL DEFAULT 0,
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
)`, `
CREATE TABLE order_payments (
	id BIGINT NOT NULL PRIMARY KEY AUTO_INCREMENT,
	order_id VARCHAR(64) NOT NULL,
	gateway VARCHAR(32) NOT NULL DEFAULT 'tetraminator',
	pay_id VARCHAR(128) NOT NULL,
	amount BIGINT NOT NULL,
	status VARCHAR(24) NOT NULL DEFAULT 'pending',
	inquiry_status VARCHAR(32) NULL,
	callback_seen BOOLEAN NOT NULL DEFAULT FALSE,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	KEY ix_order_payments_order (order_id),
	KEY ix_order_payments_pay_id (pay_id)
)`); err != nil {
		return err
	}

	// Sales plans the bots offer; amounts are in Toman.
	if err := createTable(ctx, tx, dialect, "sale_plans", `
CREATE TABLE sale_plans (
	id INTEGER PRIMARY KEY,
	admin_id INTEGER NOT NULL,
	name VARCHAR(128) NOT NULL,
	description VARCHAR(512) NULL,
	price BIGINT NOT NULL,
	duration_days INTEGER NOT NULL DEFAULT 30,
	data_limit_bytes BIGINT NOT NULL DEFAULT 0,
	ip_limit INTEGER NOT NULL DEFAULT 0,
	service_id INTEGER NULL,
	visible INTEGER NOT NULL DEFAULT 1,
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
)`, `
CREATE TABLE sale_plans (
	id BIGINT NOT NULL PRIMARY KEY AUTO_INCREMENT,
	admin_id BIGINT NOT NULL,
	name VARCHAR(128) NOT NULL,
	description VARCHAR(512) NULL,
	price BIGINT NOT NULL,
	duration_days INT NOT NULL DEFAULT 30,
	data_limit_bytes BIGINT NOT NULL DEFAULT 0,
	ip_limit INT NOT NULL DEFAULT 0,
	service_id BIGINT NULL,
	visible BOOLEAN NOT NULL DEFAULT TRUE,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	KEY ix_sale_plans_admin (admin_id, visible)
)`); err != nil {
		return err
	}

	// Buyer wallets, scoped per selling admin (reseller or main admin).
	if err := createTable(ctx, tx, dialect, "buyer_wallets", `
CREATE TABLE buyer_wallets (
	id INTEGER PRIMARY KEY,
	admin_id INTEGER NOT NULL,
	buyer_telegram_id VARCHAR(32) NOT NULL,
	balance BIGINT NOT NULL DEFAULT 0,
	updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	UNIQUE (admin_id, buyer_telegram_id)
)`, `
CREATE TABLE buyer_wallets (
	id BIGINT NOT NULL PRIMARY KEY AUTO_INCREMENT,
	admin_id BIGINT NOT NULL,
	buyer_telegram_id VARCHAR(32) NOT NULL,
	balance BIGINT NOT NULL DEFAULT 0,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	UNIQUE KEY uq_buyer_wallets (admin_id, buyer_telegram_id)
)`); err != nil {
		return err
	}

	if err := createIndex(ctx, tx, dialect, "reseller_wallet_transactions", "ix_reseller_wallet_tx_admin", []string{"admin_id", "id"}, false); err != nil {
		return err
	}
	if err := createIndex(ctx, tx, dialect, "orders", "ix_orders_admin", []string{"admin_id", "created_at"}, false); err != nil {
		return err
	}
	return createIndex(ctx, tx, dialect, "orders", "ix_orders_buyer", []string{"buyer_telegram_id"}, false)
}
