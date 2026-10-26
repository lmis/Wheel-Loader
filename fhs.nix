# A bubblewrap FHS environment with everything the game needs in the usual
# places (/usr/include, /usr/lib), so ./run.sh and go run . work unchanged.
#
#   $(nix-build fhs.nix --no-out-link)/bin/wheel-loader-fhs ./run.sh
#   nix run .#fhs -- ./run.sh                 (with flakes)
#   nix-shell -A env fhs.nix                  (an interactive shell in it)
#
# The host's OpenGL driver (/run/opengl-driver on NixOS) is used as is.
{ pkgs ? import <nixpkgs> { } }:

pkgs.buildFHSEnv {
  name = "wheel-loader-fhs";
  targetPkgs = p: with p; [
    go
    gcc
    pkg-config
    libGL
    libglvnd
    mesa
    libx11
    libxcursor
    libxrandr
    libxinerama
    libxi
    libxext
    libxrender
    libxfixes
    wayland
    wayland-scanner
    wayland-protocols
    libxkbcommon
  ];
  # raylib's build needs the headers, which live in the -dev outputs
  extraOutputsToInstall = [ "dev" ];
  # with arguments, run them (e.g. ./run.sh); without, an interactive shell
  runScript = pkgs.writeShellScript "wheel-loader-fhs-run" ''
    if [ $# -eq 0 ]; then exec bash; fi
    exec "$@"
  '';
  profile = ''
    export WHEEL_LOADER_FHS=1
  '';
}
