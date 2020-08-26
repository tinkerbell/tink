#!/usr/bin/env bash

# stops the execution if a command or pipeline has an error
set -eu

if command -v tput >/dev/null && tput setaf 1 >/dev/null 2>&1; then
	# color codes
	RED="$(tput setaf 1)"
	RESET="$(tput sgr0)"
fi

ERR="${RED:-}ERROR:${RESET:-}"

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

	local tink_password
	tink_password=$(generate_password)
	local registry_password
	registry_password=$(generate_password)
	cat <<EOF
# Network interface for Tinkerbell's network
export TINKERBELL_NETWORK_INTERFACE="$tink_interface"

# Decide on a subnet for provisioning. Tinkerbell should "own" this
# network space. Its subnet should be just large enough to be able
# to provision your hardware.
export TINKERBELL_CIDR=29

# Host IP is used by provisioner to expose different services such as
# tink, boots, etc.
#
# The host IP should the first IP in the range, and the Nginx IP
# should be the second address.
export TINKERBELL_HOST_IP=192.168.1.1

# NGINX IP is used by provisioner to serve files required for iPXE boot
export TINKERBELL_NGINX_IP=192.168.1.2

# Tink server username and password
export TINKERBELL_TINK_USERNAME=admin
export TINKERBELL_TINK_PASSWORD="$tink_password"

# Docker Registry's username and password
export TINKERBELL_REGISTRY_USERNAME=admin
export TINKERBELL_REGISTRY_PASSWORD="$registry_password"

# Legacy options, to be deleted:
export FACILITY=onprem
export ROLLBAR_TOKEN=ignored
export ROLLBAR_DISABLE=1

# logging details
export LOG_DRIVER=syslog
export LOG_OPT_SERVER_ADDRESS=tcp://192.168.1.1:514
export LOG_OPT_TAG=Tinkerbell/{{.Name}}
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
