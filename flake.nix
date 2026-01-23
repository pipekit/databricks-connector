{
  description = "Databricks Connector for Argo Workflows";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    gomod2nix.url = "github:nix-community/gomod2nix";
    gomod2nix.inputs.nixpkgs.follows = "nixpkgs";
    gomod2nix.inputs.flake-utils.follows = "flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils, gomod2nix }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        buildGoApplication = gomod2nix.legacyPackages.${system}.buildGoApplication;
      in
      {
        packages.default = buildGoApplication {
          pname = "databricks-connector";
          version = self.shortRev or "dirty";

          src = ./.;

          modules = ./gomod2nix.toml;

          subPackages = [ "cmd/databricks-connector" ];

          CGO_ENABLED = 0;

          ldflags = [
            "-s" "-w"
            "-X main.version=${self.shortRev or "dirty"}"
          ];
        };

        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            gopls
            gotools
            go-tools
            goreleaser
            golangci-lint
            gomod2nix.packages.${system}.default
          ];
        };
      }
    );
}