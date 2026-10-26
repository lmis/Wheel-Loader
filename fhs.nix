# A bubblewrap FHS environment with everything the game needs in the usual
# places (/usr/include, /usr/lib), so ./run.sh and go run . work unchanged:
#   nix-shell fhs.nix        (or, with flakes: nix develop .#fhs)
#   ./run.sh
# The host's OpenGL driver (/run/opengl-driver on NixOS) is used as is.
{ pkgs ? import <nixpkgs> { } }:

(pkgs.buildFHSEnv {
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
  runScript = "bash";
  profile = ''
    export WHEEL_LOADER_FHS=1
  '';
}).env
