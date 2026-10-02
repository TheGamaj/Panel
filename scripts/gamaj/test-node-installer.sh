#!/usr/bin/env bash
set -euo pipefail

ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

colorized_echo() { :; }

SCRIPT="$ROOT/gamaj-node.sh"

eval "$(sed -n '/^select_node_version() {$/,/^}$/p' "$SCRIPT")"
select_node_version dev
[ "$SELECTED_NODE_VERSION" = "dev" ]
select_node_version latest
[ "$SELECTED_NODE_VERSION" = "latest" ]

eval "$(sed -n '/^select_xray_core_version() {$/,/^}$/p' "$SCRIPT")"
GAMAJ_XRAY_CORE_VERSION_DEFAULT=v26.5.9
GAMAJ_XRAY_CORE_VERSION=""
select_xray_core_version
[ "$GAMAJ_XRAY_CORE_VERSION" = "v26.5.9" ]

eval "$(sed -n '/^read_node_certificate_bundle() {$/,/^}$/p' "$SCRIPT")"
CERT_FILE="$TMP/cert.pem"
CERT_KEY_FILE="$TMP/cert.key"
BUNDLE_FILE="$TMP/bundle.pem"
rm -f "$CERT_FILE" "$CERT_KEY_FILE"

printf '%s\r\n' \
    '-----BEGIN CERTIFICATE-----' \
    'certificate' \
    '-----END CERTIFICATE-----' \
    '-----BEGIN PRIVATE KEY-----' \
    'private-key' >"$BUNDLE_FILE"
printf '%s\r' '-----END PRIVATE KEY-----' >>"$BUNDLE_FILE"
read_node_certificate_bundle <"$BUNDLE_FILE"

grep -qx -- '-----END CERTIFICATE-----' "$CERT_FILE"
grep -qx -- '-----END PRIVATE KEY-----' "$CERT_KEY_FILE"
if [[ "$(uname -s)" == Linux* ]]; then
    [ "$(stat -c '%a' "$CERT_KEY_FILE")" = "600" ]
fi

# The node only stops its service before uninstall when it is actually running.
eval "$(sed -n '/^stop_gamaj_node_for_uninstall() {$/,/^}$/p' "$SCRIPT")"
down_calls=0
down_gamaj_node() { down_calls=$((down_calls + 1)); }

is_gamaj_node_up() { return 1; }
stop_gamaj_node_for_uninstall
[ "$down_calls" -eq 0 ]

is_gamaj_node_up() { return 0; }
stop_gamaj_node_for_uninstall
[ "$down_calls" -eq 1 ]

# The installer must refuse anything other than binary mode.
eval "$(sed -n '/^normalize_install_mode() {$/,/^}$/p' "$SCRIPT")"
[ "$(normalize_install_mode binary)" = "binary" ]
[ "$(normalize_install_mode '')" = "binary" ]
if (normalize_install_mode docker) 2>/dev/null; then
    echo "normalize_install_mode accepted docker" >&2
    exit 1
fi

echo "node installer tests passed"
