let _pkgs = import <nixpkgs> { };
in { pkgs ? import (_pkgs.fetchFromGitHub {
  owner = "NixOS";
  repo = "nixpkgs-channels";
  #branch@date: nixpkgs-unstable@2020-09-11
  rev = "6d4b93323e7f78121f8d6db6c59f3889aa1dd931";
  sha256 = "0g2j41cx2w2an5d9kkqvgmada7ssdxqz1zvjd7hi5vif8ag0v5la";
}) { } }:

with pkgs;

mkShell {
  buildInputs = [
    git
    gnumake
    gnused
    go
    jq
    nixfmt
    nodePackages.prettier
    protobuf
    protoc-gen-doc
    pythonPackages.codespell
    shfmt
    shellcheck
    vagrant
  ];
}
