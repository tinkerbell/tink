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
    #branch@date: nixpkgs-unstable@2022-09-02
    rev = "ee01de29d2f58d56b1be4ae24c24bd91c5380cea";
    sha256 = "0829fqp43cp2ck56jympn5kk8ssjsyy993nsp0fjrnhi265hqps7";
  }) { };

  go_1_18_5 = pkgs.go;

in mkShell {
  buildInputs = [
    docker-compose
    git
    gnumake
    gnused
    go_1_18_5
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
