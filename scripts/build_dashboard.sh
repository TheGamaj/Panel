#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
# The tutorial pages are Hugo output copied into the dashboard's public assets,
# so a build without them ships a dashboard whose /tutorial-content/* 404s.
# Skipping is therefore opt-in and loud, never silent.
if [[ "${GAMAJ_SKIP_TUTORIALS:-0}" == "1" ]]; then
    echo "WARNING: GAMAJ_SKIP_TUTORIALS=1 - building without tutorial content." >&2
    echo "WARNING: /tutorial-content/* will return 404 in this build." >&2
else
    bash "$ROOT_DIR/scripts/build_tutorials.sh"
fi
cd "$ROOT_DIR/dashboard"
export MSYS_NO_PATHCONV=1
VITE_BASE_API=/api/ npm run build
cp ./build/index.html ./build/404.html
