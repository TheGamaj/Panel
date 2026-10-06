#!/usr/bin/env bash
# One-step build for the whole Gamaj Panel: tutorial content, the dashboard
# bundle, the embedded copy inside the gateway, the Go gateway and the CLI.
#
# Why this exists: `go build ./cmd/gamaj_gateway` on its own produces a binary
# with NO dashboard. The embedded tree (internal/gateway/static/dashboard/build)
# is gitignored and a fresh checkout only carries its .gitkeep, so the gateway
# starts and answers /dashboard/login with 404. Use this script - or run the
# two steps it calls - to get a runnable Panel from source.
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

# Preflight: the tutorial pages are Hugo output, and a build without them ships
# a dashboard whose /tutorial-content/* 404s. Fail here with instructions rather
# than part-way through, and let a local UI build opt out explicitly.
if [[ "${GAMAJ_SKIP_TUTORIALS:-0}" != "1" ]] && ! command -v "${HUGO_BIN:-hugo}" >/dev/null 2>&1; then
    printf '%s\n' \
        "Gamaj Panel build needs Hugo to render the tutorial pages." \
        "" \
        "  Install it:       https://gohugo.io/installation/" \
        "  Or point at it:   HUGO_BIN=/path/to/hugo bash scripts/build_all.sh" \
        "  Or build UI only: GAMAJ_SKIP_TUTORIALS=1 bash scripts/build_all.sh" \
        "                    (/tutorial-content/* will then return 404)" >&2
    exit 1
fi

# build_dashboard.sh produces dashboard/build; build_binary.sh copies that into
# the embed tree and compiles the gateway plus the CLI.
bash "$ROOT_DIR/scripts/build_dashboard.sh"
bash "$ROOT_DIR/scripts/build_binary.sh"

echo
echo "Gamaj Panel built."
echo "  gateway : dist/gamaj-server   (dist/gamaj-server.exe on Windows)"
echo "  cli     : dist/gamaj-cli      (dist/gamaj-cli.exe on Windows)"
echo "  assets  : dashboard/build -> internal/gateway/static/dashboard/build"
echo
echo "Verify it serves the dashboard:"
echo "  GAMAJ_ENV_FILE=.env ./dist/gamaj-server"
echo "  curl -o /dev/null -w '%{http_code}\\n' http://127.0.0.1:616/dashboard/login   # expect 200"
