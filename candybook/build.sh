#!/usr/bin/env bash
# Builds the single-file page: candybook/index.html
set -euo pipefail
cd "$(dirname "$0")"
{
  cat head.html
  cat body.html
  echo '<script>'
  # painted pictures from art/ ride along inside the page as data URIs
  echo 'const ART = {'
  for f in art/*.jpg art/*.png art/*.webp; do
    [ -e "$f" ] || continue
    n=$(basename "$f"); n=${n%.*}
    case "$f" in *.png) mime=image/png ;; *.webp) mime=image/webp ;; *) mime=image/jpeg ;; esac
    printf "  '%s': 'data:%s;base64,%s',\n" "$n" "$mime" "$(base64 -w0 "$f")"
  done
  echo '};'
  cat src/*.js
  echo '</script>'
} > index.html
echo "wrote $(pwd)/index.html ($(wc -c < index.html) bytes)"
