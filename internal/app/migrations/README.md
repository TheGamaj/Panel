# Migrations

This package owns Gamaj's Go database migration runtime.

It provides goose-based migrations, dialect-aware schema helpers, legacy
schema-revision detection (older Gamaj installs), migration status/version
helpers, and startup migration execution.

Downgrades are intentionally unsupported. Migration code should be idempotent
and safe across SQLite, MySQL, and MariaDB.
