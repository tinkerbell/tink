#!/usr/bin/env nix-shell
#!nix-shell -i bash
# shellcheck shell=bash

set -eux

failed=0

if ! git ls-files | xargs codespell -q 3 -I .codespell-whitelist; then
	failed=1
fi

if ! git ls-files '*.yml' '*.json' '*.md' | xargs prettier --check; then
	failed=1
fi

if ! git ls-files '*.sh' | xargs shfmt -l -d; then
	failed=1
fi

if ! git ls-files '*.sh' | xargs shellcheck; then
	failed=1
fi

exit "$failed"
