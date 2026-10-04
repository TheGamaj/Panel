#!/usr/bin/env bash
#
# Regression test for the binary release asset resolution in
# scripts/gamaj/gamaj.sh, including the no-matching-asset error path.
#
# The GitHub API is stubbed with a `curl` shell function, so this runs anywhere
# bash and jq are available: no Linux, no systemd, no network.
set -euo pipefail

SCRIPT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../gamaj" && pwd)/gamaj.sh"
[ -f "$SCRIPT" ] || {
    echo "gamaj.sh not found at $SCRIPT" >&2
    exit 1
}
command -v jq >/dev/null 2>&1 || {
    echo "jq is required for this test" >&2
    exit 1
}

fail() {
    echo "FAIL: $*" >&2
    exit 1
}

# Minimal environment for the extracted functions. Keep the message text so the
# gap report can be asserted.
colorized_echo() { printf '%s\n' "$*"; }
GAMAJ_RELEASE_REPO="TheGamaj/Panel"
GAMAJ_SUPPORTED_ARCHES="amd64 arm64"

# Pull in only the functions under test, like scripts/gamaj/test-node-installer.sh.
eval "$(sed -n '/^report_release_asset_gap() {$/,/^}$/p' "$SCRIPT")"
eval "$(sed -n '/^get_binary_release_asset_metadata() {$/,/^}$/p' "$SCRIPT")"

STUB_RELEASE_PAYLOAD=""
curl() { printf '%s' "$STUB_RELEASE_PAYLOAD"; }

# 1. A release carrying the canonical package asset resolves to archive mode.
STUB_RELEASE_PAYLOAD='{"tag_name":"is.0.0.1","assets":[{"name":"gamaj-linux-amd64.tar.gz","browser_download_url":"https://example.test/gamaj-linux-amd64.tar.gz"}]}'
out=$(get_binary_release_asset_metadata is.0.0.1 amd64)
[ "$out" = "archive|is.0.0.1|https://example.test/gamaj-linux-amd64.tar.gz|" ] \
    || fail "unexpected archive resolution: $out"

# 2. A release carrying the split server/cli assets resolves to split mode.
STUB_RELEASE_PAYLOAD='{"tag_name":"is.0.0.1","assets":[{"name":"gamaj-server-is.0.0.1-linux-amd64","browser_download_url":"https://example.test/server"},{"name":"gamaj-cli-is.0.0.1-linux-amd64","browser_download_url":"https://example.test/cli"}]}'
out=$(get_binary_release_asset_metadata is.0.0.1 amd64)
[ "$out" = "split|is.0.0.1|https://example.test/server|https://example.test/cli" ] \
    || fail "unexpected split resolution: $out"

# 3. The no-matching-asset error path must abort with a precise explanation.
STUB_RELEASE_PAYLOAD='{"tag_name":"is.0.0.1","assets":[{"name":"gamaj-linux-s390x.tar.gz","browser_download_url":"https://example.test/s390x"}]}'
set +e
gap_output=$( (get_binary_release_asset_metadata is.0.0.1 amd64) 2>&1 )
gap_status=$?
set -e
[ "$gap_status" -eq 1 ] || fail "no-matching-asset path returned $gap_status, want 1"
printf '%s' "$gap_output" | grep -q "No Gamaj binary asset matches linux-amd64" \
    || fail "gap report missing the expected explanation: $gap_output"
printf '%s' "$gap_output" | grep -q "Expected an asset named: gamaj-linux-amd64.tar.gz" \
    || fail "gap report missing the expected asset name: $gap_output"
printf '%s' "$gap_output" | grep -q -- "- gamaj-linux-s390x.tar.gz" \
    || fail "gap report did not list the published assets: $gap_output"

# 4. A release that publishes no assets at all also aborts.
STUB_RELEASE_PAYLOAD='{"tag_name":"is.0.0.1","assets":[]}'
set +e
( get_binary_release_asset_metadata is.0.0.1 amd64 ) >/dev/null 2>&1
gap_status=$?
set -e
[ "$gap_status" -eq 1 ] || fail "empty-asset release returned $gap_status, want 1"

echo "binary release asset tests passed"
