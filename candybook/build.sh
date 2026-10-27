#!/usr/bin/env bash
# Builds the single-file page: candybook/index.html
set -euo pipefail
cd "$(dirname "$0")"
{
  cat head.html
  cat body.html
  echo '<script>'
  cat src/*.js
  echo '</script>'
} > index.html
echo "wrote $(pwd)/index.html ($(wc -c < index.html) bytes)"
