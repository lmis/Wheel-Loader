{
  description = "Wheel Loader: a 3D construction-site game for two small kids";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";

  outputs = { self, nixpkgs }:
    let
      systems = [ "x86_64-linux" "aarch64-linux" ];
      forAll = f: nixpkgs.lib.genAttrs systems (system: f nixpkgs.legacyPackages.${system});
    in
    {
      # nix develop          (Nix shell)
      # nix develop .#fhs     (bubblewrap FHS env)
      devShells = forAll (pkgs: {
        default = import ./shell.nix { inherit pkgs; };
        fhs = import ./fhs.nix { inherit pkgs; };
      });
    };
}
