let _pkgs = import <nixpkgs> { };
in { pkgs ? import (_pkgs.fetchFromGitHub {
  owner = "NixOS";
  repo = "nixpkgs";
  #branch@date: 21.11@2021-12-02
  rev = "21.11";
  sha256 = "sha256-AjhmbT4UBlJWqxY0ea8a6GU2C2HdKUREkG43oRr3TZg=";
}) { } }:

with pkgs;

mkShell {
  buildInputs = [
    docker-compose
    git
    gnumake
    gnused
    jq
    nixfmt
    nodePackages.prettier
    protobuf
    python3Packages.codespell
    shellcheck
    shfmt
    vagrant
  ];
}
