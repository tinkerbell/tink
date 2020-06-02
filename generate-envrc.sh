#!/usr/bin/env bash

# stops the execution if a command or pipeline has an error
set -e

if command -v tput >>/dev/null; then
	# color codes
	RED="$(tput setaf 1)"
	RESET="$(tput sgr0)"
else
	echo "color coding will not happen as tput command not found."
fi

ERR="${RED}ERROR:$RESET"

err() (
	if [ -z "${1:-}" ]; then
		cat >&2
	else
		echo "$ERR " "$@" >&2
	fi
)

candidate_interfaces() (
	ip -o link show |
		awk -F': ' '{print $2}' |
		sed 's/[ \t].*//;/^\(lo\|bond0\|\|\)$/d' |
		sort
)

validate_tinkerbell_network_interface() (
	local tink_interface=$1

	if ! candidate_interfaces | grep -q "^$tink_interface$"; then
		err "Invalid interface ($tink_interface) selected, must be one of:"
		candidate_interfaces | err
		return 1
	else
		return 0
	fi
)

generate_password() (
	head -c 12 /dev/urandom | sha256sum | cut -d' ' -f1
)

generate_envrc() (
	local tink_interface=$1

	validate_tinkerbell_network_interface "$tink_interface"

	local registry_password
	registry_password=$(generate_password)
	cat <<EOF
# Network interface for Tinkerbell's network
export TINKERBELL_NETWORK_INTERFACE=$tink_interface"

# Subnet (IP block) used by Tinkerbell's provisioning tools
# Hint: calculate the values in this file with ipcalc:
#
# $ ipcalc 192.168.1.0/29
export TINKERBELL_NETWORK=192.168.1.0/29

export TINKERBELL_CIDR=29

# Host IP is used by provisioner to expose different services such as tink, boots, etc.
export TINKERBELL_HOST_IP=192.168.1.1

# NGINX IP is used by provisioner to serve files required for iPXE boot
export TINKERBELL_NGINX_IP=192.168.1.2

# Netmask for Tinkerbell network
export TINKERBELL_NETMASK=255.255.255.248

# Broadcast IP for Tinkerbell network
export TINKERBELL_BROADCAST_IP=192.168.1.7

# Docker Registry's username and password
export TINKERBELL_REGISTRY_USERNAME=admin
export TINKERBELL_REGISTRY_PASSWORD=$registry_password

# Legacy options, to be deleted:
export FACILITY=onprem
export ROLLBAR_TOKEN=ignored
export ROLLBAR_DISABLE=1
EOF
)

main() (
	if [ -z "${1:-}" ]; then
		err "Usage: $0 network-interface-name > envrc"
		exit 1
	fi

	generate_envrc "$1"
)

main "$@"
