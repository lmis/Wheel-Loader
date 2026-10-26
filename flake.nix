{
  description = "Wheel Loader: a 3D construction-site game for two small kids";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";

  outputs = { self, nixpkgs }:
    let
      systems = [ "x86_64-linux" "aarch64-linux" ];
      forAll = f: nixpkgs.lib.genAttrs systems (system: f nixpkgs.legacyPackages.${system});
    in
    {
      # nix develop                  (Nix shell)
      # nix develop .#fhs             (a shell in the bubblewrap FHS env)
      devShells = forAll (pkgs: {
        default = import ./shell.nix { inherit pkgs; };
        fhs = (import ./fhs.nix { inherit pkgs; }).env;
      });
      # nix run .#fhs -- ./run.sh     (run the game in the FHS env)
      packages = forAll (pkgs: { fhs = import ./fhs.nix { inherit pkgs; }; });
      apps = forAll (pkgs: {
        fhs = {
          type = "app";
          program = "${import ./fhs.nix { inherit pkgs; }}/bin/wheel-loader-fhs";
        };
      });
    };
}
