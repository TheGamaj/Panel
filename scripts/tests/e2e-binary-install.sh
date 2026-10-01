#!/usr/bin/env bash
#
# End-to-end check for the Gamaj binary install path.
#
# It builds the dashboard and the Go binaries from this checkout, runs the real
# installer functions against a sandbox prefix, starts the panel as a systemd
# service, and fails when the service is not active or the panel does not answer
# on its gateway port (616 by default).
#
# The sandbox keeps the host clean: the install prefix, the data root, the CLI
# launcher and the script path all live under a temporary directory, and the
# temporary systemd unit is removed again on exit.
#
# Requirements: Linux, systemd, root (run through sudo), Go, Node.js and npm.
# On any other platform the test reports SKIP and succeeds, so it never blocks
# Windows or macOS development.
set -euo pipefail

SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
REPO_ROOT=$(cd "$SCRIPT_DIR/../.." && pwd)

E2E_PREFIX="${GAMAJ_E2E_PREFIX:-$(mktemp -d /tmp/gamaj-e2e-XXXXXX)}"
E2E_APP_NAME="${GAMAJ_E2E_APP_NAME:-gamaj-e2e}"
E2E_PANEL_PORT="${GAMAJ_E2E_PANEL_PORT:-616}"
E2E_WAIT_SECONDS="${GAMAJ_E2E_WAIT_SECONDS:-120}"
E2E_ADMIN_USER="${GAMAJ_E2E_ADMIN_USER:-e2e-admin}"
E2E_ADMIN_PASSWORD="${GAMAJ_E2E_ADMIN_PASSWORD:-e2e-password-123}"

PANEL_UNIT="/etc/systemd/system/$E2E_APP_NAME.service"
PANEL_URL="http://127.0.0.1:$E2E_PANEL_PORT"
STUB_BIN="$E2E_PREFIX/stub-bin"

skip() {
    echo "SKIP: $*"
    exit 0
}

step() {
    echo
    echo "== $*"
}

fail() {
    echo "FAIL: $*" >&2
    exit 1
}

cleanup() {
    local exit_code=$?
    if [ "${E2E_SOURCED:-0}" = "1" ]; then
        systemctl disable --now "$E2E_APP_NAME.service" >/dev/null 2>&1 || true
    fi
    if [ -f "$PANEL_UNIT" ]; then
        rm -f "$PANEL_UNIT"
        systemctl daemon-reload >/dev/null 2>&1 || true
    fi
    if [ "${GAMAJ_E2E_KEEP_PREFIX:-0}" = "1" ]; then
        echo "Kept sandbox at $E2E_PREFIX"
    else
        rm -rf "$E2E_PREFIX"
    fi
    exit "$exit_code"
}
trap cleanup EXIT

step "Check prerequisites"
[[ "$(uname -s)" == Linux* ]] || skip "the binary install path is Linux-only"
[ -d /run/systemd/system ] || skip "systemd is not running"
command -v systemctl >/dev/null 2>&1 || skip "systemctl is not available"
[ "$(id -u)" -eq 0 ] || skip "the installer needs root (run this script with sudo)"
command -v go >/dev/null 2>&1 || skip "the Go toolchain is required to build the binaries"
command -v npm >/dev/null 2>&1 || skip "npm is required to build the dashboard"
command -v curl >/dev/null 2>&1 || skip "curl is required"

export GAMAJ_INSTALL_DIR="$E2E_PREFIX/opt"
export GAMAJ_DATA_ROOT="$E2E_PREFIX/var/lib"
export GAMAJ_CLI_LAUNCHER="$GAMAJ_INSTALL_DIR/bin/gamaj-cli"
export GAMAJ_SCRIPT_INSTALL_PATH="$GAMAJ_INSTALL_DIR/bin/gamaj"
APP_DIR="$GAMAJ_INSTALL_DIR/$E2E_APP_NAME"
DATA_DIR="$GAMAJ_DATA_ROOT/$E2E_APP_NAME"

step "Build dashboard and binaries"
if [ ! -f "$REPO_ROOT/dashboard/build/index.html" ]; then
    ( cd "$REPO_ROOT/dashboard" && npm ci --no-audit --no-fund >/dev/null && VITE_BASE_API=/api/ npm run build >/dev/null && cp build/index.html build/404.html )
fi
[ -f "$REPO_ROOT/dashboard/build/index.html" ] || fail "dashboard build did not produce dashboard/build/index.html"
( cd "$REPO_ROOT" && bash scripts/build_binary.sh >/dev/null )
[ -x "$REPO_ROOT/dist/gamaj-server" ] || fail "dist/gamaj-server was not built"
[ -x "$REPO_ROOT/dist/gamaj-cli" ] || fail "dist/gamaj-cli was not built"

step "Prepare sandbox configuration"
mkdir -p "$APP_DIR" "$DATA_DIR" "$STUB_BIN"
# The installer only installs packages that are missing. certbot is not part of
# the panel runtime, so a stub keeps the test from pulling the ACME toolchain.
printf '#!/bin/sh\nexit 0\n' > "$STUB_BIN/certbot"
chmod +x "$STUB_BIN/certbot"
export PATH="$STUB_BIN:$PATH"

{
    echo "[TEMPLATE]"
    echo "GAMAJ_DATABASE_URL=sqlite:///$DATA_DIR/db.sqlite3"
    echo "GAMAJ_DATABASE_FLAVOR=sqlite"
    echo "GAMAJ_HOST=127.0.0.1"
    echo "GAMAJ_DATA_DIR=$DATA_DIR"
    if [ "$E2E_PANEL_PORT" = "616" ]; then
        # Leave the gateway port unset so the test proves the built-in default.
        echo "# gateway port uses the Gamaj default: 616"
    else
        echo "GAMAJ_PORT=$E2E_PANEL_PORT"
    fi
} > "$APP_DIR/.env"

step "Install the panel through the installer functions"
export GAMAJ_APP_NAME="$E2E_APP_NAME"
export GAMAJ_BINARY_SERVER_OVERRIDE="$REPO_ROOT/dist/gamaj-server"
export GAMAJ_BINARY_CLI_OVERRIDE="$REPO_ROOT/dist/gamaj-cli"
export GAMAJ_BINARY_OVERRIDE_VERSION="is.0.0.1"
export GAMAJ_SOURCE_ONLY=1
E2E_SOURCED=1

# shellcheck disable=SC1091
source "$REPO_ROOT/scripts/gamaj/gamaj.sh"

if normalize_install_mode docker >/dev/null 2>&1; then
    fail "the installer still accepts the removed docker install mode"
fi
[ "$(get_install_mode)" = "binary" ] || fail "the installer did not report binary install mode"

install_binary_gamaj latest sqlite 0
up_gamaj

grep -qi docker "$PANEL_UNIT" && fail "$E2E_APP_NAME.service still references docker"

step "Migrate the database and create an admin"
gamaj_cli migrate up
gamaj_cli admin create "$E2E_ADMIN_USER" --role full_access --password "$E2E_ADMIN_PASSWORD"

step "Wait for the panel on port $E2E_PANEL_PORT"
healthy=0
for _ in $(seq 1 "$E2E_WAIT_SECONDS"); do
    if curl -fsS --max-time 3 "$PANEL_URL/__gamaj_go/healthz" >/dev/null 2>&1; then
        healthy=1
        break
    fi
    if ! systemctl is-active --quiet "$E2E_APP_NAME.service"; then
        journalctl -u "$E2E_APP_NAME.service" --no-pager -n 50 || true
        fail "$E2E_APP_NAME.service stopped before the panel became healthy"
    fi
    sleep 1
done
[ "$healthy" -eq 1 ] || {
    journalctl -u "$E2E_APP_NAME.service" --no-pager -n 50 || true
    fail "the panel did not answer on $PANEL_URL within ${E2E_WAIT_SECONDS}s"
}

step "Assert the gateway port and the dashboard"
ss -ltn | awk '{print $4}' | grep -qE "[:.]$E2E_PANEL_PORT\$" || fail "nothing is listening on port $E2E_PANEL_PORT"
curl -fsS --max-time 5 "$PANEL_URL/" | grep -qi 'id="root"' || fail "the dashboard HTML was not served on port $E2E_PANEL_PORT"

step "Assert admin login through the API"
login_body=$(curl -fsS --max-time 10 -H 'Content-Type: application/json' \
    -d "{\"username\":\"$E2E_ADMIN_USER\",\"password\":\"$E2E_ADMIN_PASSWORD\"}" \
    "$PANEL_URL/api/admin/token") || fail "admin login failed"
echo "$login_body" | grep -q 'access_token' || fail "admin login did not return an access token"

echo
echo "PASS: Gamaj binary install is healthy on port $E2E_PANEL_PORT"
