{
  description = "Terraform Provider for Technitium DNS Server";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            curl
            docker
            git
            go
            go-task
            golangci-lint
            gotools
            opentofu
            pre-commit
          ];

          shellHook = ''
            echo "Terraform Provider for Technitium DNS Server development environment"
            echo "Go version: $(go version)"
            echo "Terraform version: $(tofu version)"
            echo "Available tasks: $(task --list-all)"
          '';
        };
      });
}
