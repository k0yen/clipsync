{
  description = "ClipSync: A privacy-first realtime shared clipboard";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixpkgs-unstable";
    # Use the explicit github: prefix to avoid registry lookups
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        devShells.default = pkgs.mkShell {
          nativeBuildInputs = with pkgs; [
            go
            bun
            air
          ];
          shellHook = ''
            echo "--- ClipSync Flake Environment Active ---"
          '';
        };
      });
}

