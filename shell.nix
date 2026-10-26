# Development shell with everything needed to build and run the game:
#   nix-shell          (or, with flakes: nix develop)
#   ./run.sh           (or go run .)
{ pkgs ? import <nixpkgs> { } }:

let
  # libraries raylib (GLFW) compiles against and loads at run time
  libs = with pkgs; [
    libGL
    libx11
    libxcursor
    libxrandr
    libxinerama
    libxi
    libxext
    libxrender
    libxfixes
    wayland
    wayland-protocols
    libxkbcommon
  ];
in
pkgs.mkShell {
  packages = with pkgs; [ go gcc pkg-config wayland-scanner ] ++ libs;
  # GLFW opens libGL, X11 and Wayland with dlopen, so put them on the path.
  # The OpenGL driver itself comes from the system on NixOS
  # (/run/opengl-driver); elsewhere Mesa is the fallback (Intel and AMD GPUs,
  # or software rendering).
  shellHook = ''
    export LD_LIBRARY_PATH="${pkgs.lib.makeLibraryPath libs}''${LD_LIBRARY_PATH:+:$LD_LIBRARY_PATH}"
    if [ -d /run/opengl-driver/lib ]; then
      export LD_LIBRARY_PATH="/run/opengl-driver/lib:$LD_LIBRARY_PATH"
    else
      export LD_LIBRARY_PATH="$LD_LIBRARY_PATH:${pkgs.mesa}/lib"
      export LIBGL_DRIVERS_PATH="${pkgs.mesa}/lib/dri"
      export __EGL_VENDOR_LIBRARY_DIRS="${pkgs.mesa}/share/glvnd/egl_vendor.d"
    fi
    echo "wheel-loader dev shell: go $(go env GOVERSION | sed s/go//). Run ./run.sh or go run ."
  '';
}
