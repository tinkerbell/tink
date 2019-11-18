#!/usr/bin/env sh

# set -o errexit -o nounset -o pipefail

if [ -z "${ROVER_TLS_CERT:-}" ]; then
	(
		FACILITY=$(echo "$FACILITY" | tr '[:upper:]' '[:lower:]')
		echo "creating directory"
		mkdir -p "certs"
		FACILITY=$FACILITY sh gencerts.sh
		rm server.csr server-csr.json
		rm ca.csr ca.json
	)
fi

"$@"
