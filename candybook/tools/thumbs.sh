#!/usr/bin/env bash
# Rebuild the button pictures from the book itself, then rebuild the book.
#   tools/thumbs.sh            every page
#   tools/thumbs.sh morning    only some pages
set -euo pipefail
cd "$(dirname "$0")/.."
raw=$(mktemp -d)
./build.sh >/dev/null
node tools/capture.js "$raw" "$@"
python3 tools/crop.py "$raw"
rm -rf "$raw"
./build.sh
