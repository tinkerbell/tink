#!/usr/bin/env sh

# set -o errexit -o nounset -o pipefail

if [ -z "${TINKERBELL_TLS_CERT:-}" ]; then
	(
		echo "creating directory"
		mkdir -p "certs"
		./gencerts.sh
	)
fi

"$@"
