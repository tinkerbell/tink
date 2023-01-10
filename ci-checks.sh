#!/usr/bin/env nix-shell
#!nix-shell -i bash
# shellcheck shell=bash

set -eux

failed=0

# spell-checks only language files to avoid spell-checking checksums
if ! git ls-files '*.sh' '*.go' | xargs codespell -q 3 -I .codespell-whitelist; then
	failed=1
fi

# --check doesn't show what line number fails, so write the result to disk for the diff to catch
if ! git ls-files '*.json' | xargs prettier --list-different --write; then
	failed=1
fi

if ! git ls-files '*.sh' | xargs shfmt -l -d; then
	failed=1
fi

if ! git ls-files '*.sh' | xargs shellcheck; then
	failed=1
fi

if ! nixfmt shell.nix; then
	failed=1
fi

if ! git diff | (! grep .); then
	failed=1
fi

exit "$failed"
