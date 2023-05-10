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
    git
    gnumake
    jq
    nixfmt
    nodePackages.prettier
    python3Packages.codespell
    python3Packages.pip
    python3Packages.setuptools
    shellcheck
    shfmt
  ];
}
