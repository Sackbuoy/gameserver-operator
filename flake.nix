{
  description = "A basic go dev flake";
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    systems.url = "github:nix-systems/default";
    flake-utils = {
      url = "github:numtide/flake-utils";
      inputs.systems.follows = "systems";
    };
    go-flake.url = "github:Sackbuoy/flakes?dir=go/go";
    golangci-lint-flake.url = "github:Sackbuoy/flakes?dir=go/golangci-lint";
  };

  outputs = {
    nixpkgs,
    flake-utils,
    go-flake,
    golangci-lint-flake,
    ...
  }:
    flake-utils.lib.eachDefaultSystem (
      system: let
        pkgs = nixpkgs.legacyPackages.${system};

        golangciPackage = golangci-lint-flake.lib.getVersion ./.github/workflows/pr.yaml;
        goPackage = go-flake.lib.getVersion ./go.mod;
      in {
        devShells.default = pkgs.mkShell {
          packages = [
            pkgs.bashInteractive
          ];
          buildInputs = [
            golangci-lint-flake.packages.${system}.${golangciPackage}
            go-flake.packages.${system}.${goPackage}
            pkgs.delve
            pkgs.bashInteractive
            pkgs.gotools
            pkgs.gopls
          ];

          CGO_CFLAGS = "-O2";

          env = {
            GO111MODULE = "on";
          };

          shellHook = ''
            echo "Golang development environment with:"
            echo "Go: ${goPackage}"
            echo "Golangci-lint ${golangciPackage}"
          '';
        };
      }
    );
}
