let _pkgs = import <nixpkgs> { };
in { pkgs ? import (_pkgs.fetchFromGitHub {
  owner = "NixOS";
  repo = "nixpkgs-channels";
  #branch@date: nixpkgs-unstable@2020-02-01
  rev = "e3a9318b6fdb2b022c0bda66d399e1e481b24b5c";
  sha256 = "1hlblna9j0afvcm20p15f5is7cmwl96mc4vavc99ydc4yc9df62a";
}) { } }:

with pkgs;

mkShell {
  buildInputs = [
    gnumake
    go
    pythonPackages.codespell
    shfmt
    shellcheck
    terraform
  ];
}
