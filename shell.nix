let _pkgs = import <nixpkgs> { };
in { pkgs ? import (_pkgs.fetchFromGitHub {
  owner = "NixOS";
  repo = "nixpkgs";
  #branch@date: nixos-unstable-small@2022-04-18
  rev = "e33fe968df5a2503290682278399b1198f7ba56f";
  sha256 = "0kr30yj9825jx4zzcyn43c398mx3l63ndgfrg1y9v3d739mfgyw3";
}) { } }:

with pkgs;

mkShell {
  buildInputs = [
    docker-compose
    git
    gnumake
    gnused
    go_1_18
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
