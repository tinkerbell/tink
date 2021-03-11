let _pkgs = import <nixpkgs> { };
in { pkgs ? import (_pkgs.fetchFromGitHub {
  owner = "NixOS";
  repo = "nixpkgs";
  #branch@date: nixpkgs@2021-03-11
  rev = "ee9b92ddf8734c48d554a920c038c1b569a72b74";
  sha256 = "101fh81qx1nx3a87q0m26nsy6br3sjrygm2i91pv9w0fwn5s0x4k";
}) { } }:

with pkgs;

mkShell {
  buildInputs = [
    git
    gnumake
    gnused
    go
    gotools
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
