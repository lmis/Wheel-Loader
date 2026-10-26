#!/usr/bin/env bash
# Build and run Wheel Loader on Linux. Just run:   ./run.sh
#
# It installs what is missing (the OpenGL/X11 headers raylib needs, and Go
# itself if yours is too old), rebuilds the game when the code has changed,
# and starts it. Any arguments are passed to the game, e.g. ./run.sh -studio shots
set -euo pipefail

cd "$(dirname "$0")"
GO_VERSION=1.26.0
CACHE="${XDG_CACHE_HOME:-$HOME/.cache}/wheel-loader"
BIN=bin/wheel-loader

say() { printf '\033[1;33m==>\033[0m %s\n' "$*"; }
die() { printf '\033[1;31mError:\033[0m %s\n' "$*" >&2; exit 1; }

[ "$(uname -s)" = Linux ] || die "this script is for Linux; on Mac or Windows run: go run ."

sudo_() { if [ "$(id -u)" -eq 0 ]; then "$@"; else sudo "$@"; fi; }

# --- system libraries raylib compiles against -------------------------------
need_libs() {
	for h in GL/gl.h X11/Xlib.h X11/extensions/Xrandr.h X11/extensions/Xinerama.h \
		X11/Xcursor/Xcursor.h X11/extensions/XInput2.h xkbcommon/xkbcommon.h; do
		[ -e "/usr/include/$h" ] || return 0
	done
	command -v gcc >/dev/null || command -v cc >/dev/null || return 0
	return 1
}

if need_libs; then
	say "Installing the OpenGL/X11 development libraries (may ask for your password)"
	if command -v apt-get >/dev/null; then
		sudo_ apt-get update -qq
		sudo_ apt-get install -y build-essential libgl1-mesa-dev libxi-dev libxcursor-dev \
			libxrandr-dev libxinerama-dev libwayland-dev libxkbcommon-dev
	elif command -v dnf >/dev/null; then
		sudo_ dnf install -y gcc mesa-libGL-devel libXi-devel libXcursor-devel libXrandr-devel \
			libXinerama-devel wayland-devel libxkbcommon-devel
	elif command -v pacman >/dev/null; then
		sudo_ pacman -S --needed --noconfirm base-devel mesa libxi libxcursor libxrandr libxinerama \
			wayland libxkbcommon
	elif command -v zypper >/dev/null; then
		sudo_ zypper install -y gcc Mesa-libGL-devel libXi-devel libXcursor-devel libXrandr-devel \
			libXinerama-devel wayland-devel libxkbcommon-devel
	else
		die "unknown package manager; install a C compiler and the OpenGL, X11 (Xi, Xcursor, Xrandr, Xinerama), wayland and xkbcommon development packages"
	fi
fi

# --- Go ----------------------------------------------------------------------
go_ok() { # is this go at least $GO_VERSION?
	local v
	v=$("$1" env GOVERSION 2>/dev/null | sed 's/^go//') || return 1
	[ -n "$v" ] && [ "$(printf '%s\n%s\n' "$GO_VERSION" "$v" | sort -V | head -1)" = "$GO_VERSION" ]
}

GO=""
if command -v go >/dev/null && go_ok go; then
	GO=go
elif [ -x "$CACHE/go/bin/go" ] && go_ok "$CACHE/go/bin/go"; then
	GO="$CACHE/go/bin/go"
else
	case "$(uname -m)" in
	x86_64) arch=amd64 ;;
	aarch64 | arm64) arch=arm64 ;;
	*) die "unsupported CPU $(uname -m); install Go $GO_VERSION or newer yourself" ;;
	esac
	say "Downloading Go $GO_VERSION (once, into $CACHE)"
	mkdir -p "$CACHE"
	rm -rf "$CACHE/go"
	curl -fsSL "https://go.dev/dl/go$GO_VERSION.linux-$arch.tar.gz" | tar -xz -C "$CACHE"
	GO="$CACHE/go/bin/go"
fi

# --- build (only when something changed) -------------------------------------
if [ ! -x "$BIN" ] || [ -n "$(find . -path ./bin -prune -o \( -name '*.go' -o -name go.mod -o -name go.sum -o -path './assets/*' \) -newer "$BIN" -print -quit)" ]; then
	say "Building (the first build compiles raylib and takes a minute)"
	mkdir -p bin
	"$GO" build -o "$BIN" .
fi

say "Starting Wheel Loader"
exec "./$BIN" "$@"
