let _pkgs = import <nixpkgs> { };
in { pkgs ? import (_pkgs.fetchFromGitHub {
  owner = "NixOS";
  repo = "nixpkgs";
  #branch@date: 21.11@2021-12-02
  rev = "21.11";
  sha256 = "sha256-AjhmbT4UBlJWqxY0ea8a6GU2C2HdKUREkG43oRr3TZg=";
}) { } }:

with pkgs;

let
  pkgs = import (_pkgs.fetchFromGitHub {
    # go 1.18.5
    owner = "NixOS";
    repo = "nixpkgs";
    # branch@date: nixpkgs-unstable@2023-03-30
    rev = "8b3bc690e201c8d3cbd14633dbf3462a820e73f2";
    sha256 = "sha256-+ckiCxbGFSs1/wHKCXAZnvb37Htf6k5nmQE3T0Y7hK8=";
  }) { };

  go_1_20_3 = pkgs.go;

in mkShell {
  buildInputs = [
    docker-compose
    git
    gnumake
    gnused
    go_1_20_3
    jq
    nixfmt
    nodePackages.prettier
    protobuf
    python3Packages.codespell
    python3Packages.pip
    python3Packages.setuptools
    shellcheck
    shfmt
  ];
}
