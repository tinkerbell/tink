#!/usr/bin/env nix-shell
#!nix-shell -i bash

set -eux

codespell -q 3 -I .codespell-whitelist *
prettier --check '**/*.json' '**/*.md' '**/*.yml'
shfmt -l -d .
