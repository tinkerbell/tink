#!/bin/bash

generate_certificates() (
	mkdir -p "$STATEDIR/certs"

	if [ ! -f "$STATEDIR/certs/ca.json" ]; then
		jq \
			'.
			 | .names[0].L = $facility
			' \
			"$DEPLOYDIR/tls/ca.in.json" \
			--arg ip "$TINKERBELL_HOST_IP" \
			--arg facility "$FACILITY" \
			>"$STATEDIR/certs/ca.json"
	fi

	if [ ! -f "$STATEDIR/certs/server-csr.json" ]; then
		jq \
			'.
			| .hosts += [ $ip, "tinkerbell.\($facility).packet.net" ]
			| .names[0].L = $facility
			| .hosts = (.hosts | sort | unique)
			' \
			"$DEPLOYDIR/tls/server-csr.in.json" \
			--arg ip "$TINKERBELL_HOST_IP" \
			--arg facility "$FACILITY" \
			>"$STATEDIR/certs/server-csr.json"
	fi

	docker build --tag "tinkerbell-certs" "$DEPLOYDIR/tls"
	docker run --rm \
		--volume "$STATEDIR/certs:/certs" \
		--user "$UID:$(id -g)" \
		tinkerbell-certs

	local certs_dir="/etc/docker/certs.d/$TINKERBELL_HOST_IP"

	# copy public key to NGINX for workers
	#if ! cmp --quiet "$STATEDIR"/certs/ca.pem "$STATEDIR/webroot/workflow/ca.pem"; then
	#	cp "$STATEDIR"/certs/ca.pem "$STATEDIR/webroot/workflow/ca.pem"
	#fi

	# update host to trust registry certificate
	if ! cmp --quiet "$STATEDIR/certs/ca.pem" "$certs_dir/tinkerbell.crt"; then
		if [ ! -d "$certs_dir/tinkerbell.crt" ]; then
			# The user will be told to create the directory
			# in the next block, if copying the certs there
			# fails.
			sudo mkdir -p "$certs_dir" || true >/dev/null 2>&1
		fi
		if ! sudo cp "$STATEDIR/certs/ca.pem" "$certs_dir/tinkerbell.crt"; then
			echo "$ERR please copy $STATEDIR/certs/ca.pem to $certs_dir/tinkerbell.crt"
			echo "$BLANK and run $0 again:"

			if [ ! -d "$certs_dir" ]; then
				echo "sudo mkdir -p '$certs_dir'"
			fi
			echo "sudo cp '$STATEDIR/certs/ca.pem' '$certs_dir/tinkerbell.crt'"

			exit 1
		fi
	fi
)

DEPLOYDIR=${PWD//test/deploy}
STATEDIR="$DEPLOYDIR"/state
generate_certificates
