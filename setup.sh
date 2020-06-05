#!/usr/bin/env bash

# stops the execution if a command or pipeline has an error
set -e

# Tinkerbell stack Linux setup script
#
# See https://tinkerbell.org/setup for the installation steps.

# file to hold all environment variables
ENV_FILE=envrc

SCRATCH=$(mktemp -d -t tmp.XXXXXXXXXX)
readonly SCRATCH
function finish() (
	rm -rf "$SCRATCH"
)
trap finish EXIT

DEPLOYDIR=$(pwd)/deploy
readonly DEPLOYDIR

if command -v tput >>/dev/null; then
	# color codes
	RED="$(tput setaf 1)"
	GREEN="$(tput setaf 2)"
	YELLOW="$(tput setaf 3)"
	RESET="$(tput sgr0)"
else
	echo "color coding will not happen as tput command not found."
fi

INFO="${GREEN}INFO:$RESET"
ERR="${RED}ERROR:$RESET"
WARN="${YELLOW}WARNING:$RESET"
BLANK="      "
NEXT="${GREEN}NEXT:$RESET"

get_distribution() (
	local lsb_dist=""
	# Every system that we officially support has /etc/os-release
	if [ -r /etc/os-release ]; then
		# shellcheck disable=SC1091
		lsb_dist="$(. /etc/os-release && echo "$ID")"
	fi
	# Returning an empty string here should be alright since the
	# case statements don't act unless you provide an actual value
	echo "$lsb_dist"
)

get_distro_version() (
	local lsb_version="0"
	# Every system that we officially support has /etc/os-release
	if [ -r /etc/os-release ]; then
		# shellcheck disable=SC1091
		lsb_version="$(. /etc/os-release && echo "$VERSION_ID")"
	fi

	echo "$lsb_version"
)

is_network_configured() (
	# Require the provisioner interface have both the host and nginx IP
	if ! ip addr show "$TINKERBELL_NETWORK_INTERFACE" |
		grep -q "$TINKERBELL_HOST_IP"; then
		return 1
	fi

	if ! ip addr show "$TINKERBELL_NETWORK_INTERFACE" |
		grep -q "$TINKERBELL_NGINX_IP"; then
		return 1
	fi

	return 0
)

setup_networking() (
	distro=$1
	version=$2

	setup_network_forwarding

	if is_network_configured; then
		echo "$INFO tinkerbell network interface is already configured"
		return 0
	fi

	case "$distro" in
	ubuntu)
		if (($(echo "$version >= 17.10" | bc -l))); then
			setup_networking_netplan
		else
			setup_networking_ubuntu_legacy
		fi
		;;
	centos)
		setup_networking_centos
		;;
	*)
		echo "$ERR this setup script cannot configure $distro ($version)"
		echo "$BLANK please read this script's source and configure it manually."
		exit 1
		;;
	esac
	if is_network_configured; then
		echo "$INFO tinkerbell network interface configured successfully"
	else
		echo "$ERR tinkerbell network interface configuration failed"
	fi
)

setup_network_forwarding() (
	# enable IP forwarding for docker
	if [ "$(sysctl -n net.ipv4.ip_forward)" != "1" ]; then
		if [ -d /etc/sysctl.d ]; then
			echo "net.ipv4.ip_forward=1" >/etc/sysctl.d/99-tinkerbell.conf
		elif [ -f /etc/sysctl.conf ]; then
			echo "net.ipv4.ip_forward=1" >>/etc/sysctl.conf
		fi

		sysctl net.ipv4.ip_forward=1
	fi
)

setup_networking_netplan() (
	jq -n \
		--arg interface "$TINKERBELL_NETWORK_INTERFACE" \
		--arg cidr "$TINKERBELL_CIDR" \
		--arg host_ip "$TINKERBELL_HOST_IP" \
		--arg nginx_ip "$TINKERBELL_NGINX_IP" \
		'{
  network: {
    renderer: "networkd",
    ethernets: {
      ($interface): {
        addresses: [
          "\($host_ip)/\($cidr)",
          "\($nginx_ip)/\($cidr)"
        ]
      }
    }
  }
}' >"/etc/netplan/${TINKERBELL_NETWORK_INTERFACE}.yaml"

	ip link set "$TINKERBELL_NETWORK_INTERFACE" nomaster
	netplan apply
	echo "$INFO waiting for the network configuration to be applied by systemd-networkd"
	sleep 3
)

setup_networking_ubuntu_legacy() (
	if [ ! -f /etc/network/interfaces ]; then
		echo "$ERR file /etc/network/interfaces not found"
		exit 1
	fi

	if grep -q "$TINKERBELL_NETWORK_INTERFACE" /etc/network/interfaces; then
		echo "$ERR /etc/network/interfaces already has an entry for $TINKERBELL_NETWORK_INTERFACE."
		echo "$BLANK To prevent breaking your network, please edit /etc/network/interfaces"
		echo "$BLANK and configure $TINKERBELL_NETWORK_INTERFACE as follows:"
		generate_iface_config
		echo ""
		echo "$BLANK Then run the following commands:"
		echo "$BLANK ip link set $TINKERBELL_NETWORK_INTERFACE nomaster"
		echo "$BLANK ifdown $TINKERBELL_NETWORK_INTERFACE:0"
		echo "$BLANK ifdown $TINKERBELL_NETWORK_INTERFACE:1"
		echo "$BLANK ifup $TINKERBELL_NETWORK_INTERFACE:0"
		echo "$BLANK ifup $TINKERBELL_NETWORK_INTERFACE:1"
		exit 1
	else
		generate_iface_config >>/etc/network/interfaces
		ip link set "$TINKERBELL_NETWORK_INTERFACE" nomaster
		ifdown "$TINKERBELL_NETWORK_INTERFACE:0"
		ifdown "$TINKERBELL_NETWORK_INTERFACE:1"
		ifup "$TINKERBELL_NETWORK_INTERFACE:0"
		ifup "$TINKERBELL_NETWORK_INTERFACE:1"
	fi
)

generate_iface_config() (
	cat <<EOF

auto $TINKERBELL_NETWORK_INTERFACE:0
iface $TINKERBELL_NETWORK_INTERFACE:0 inet static
    address $TINKERBELL_HOST_IP
    netmask $TINKERBELL_NETMASK
    broadcast $TINKERBELL_BROADCAST_IP
    pre-up sleep 4

auto $TINKERBELL_NETWORK_INTERFACE:1
iface $TINKERBELL_NETWORK_INTERFACE:1 inet static
    address $TINKERBELL_NGINX_IP
    netmask $TINKERBELL_NETMASK
    broadcast $TINKERBELL_BROADCAST_IP
    pre-up sleep 4

EOF
)

setup_networking_centos() (
	local HWADDRESS
	local content

	HWADDRESS=$(ip addr show "$TINKERBELL_NETWORK_INTERFACE" | grep ether | awk -F 'ether' '{print $2}' | cut -d" " -f2)
	content=$(
		cat <<EOF
DEVICE=$TINKERBELL_NETWORK_INTERFACE
ONBOOT=yes
HWADDR=$HWADDRESS
BOOTPROTO=static

IPADDR0=$TINKERBELL_HOST_IP
NETMASK0=$TINKERBELL_NETMASK
IPADDR1=$TINKERBELL_NGINX_IP
NETMASK1=$TINKERBELL_NETMASK
EOF
	)

	local cfgfile="/etc/sysconfig/network-scripts/ifcfg-$TINKERBELL_NETWORK_INTERFACE"

	if [ -f "$cfgfile" ]; then
		echo "$ERR network config already exists: $cfgfile"
		echo "$BLANK Please update it to match this configuration:"
		echo "$content"
		echo ""
		echo "$BLANK Then, run the following commands:"
		echo "ip link set $TINKERBELL_NETWORK_INTERFACE nomaster"
		echo "ifup $TINKERBELL_NETWORK_INTERFACE"
	fi

	echo "$content" >"$cfgfile"

	ip link set "$TINKERBELL_NETWORK_INTERFACE" nomaster
	ifup "$TINKERBELL_NETWORK_INTERFACE"
)

setup_osie() (
	mkdir -p "$DEPLOYDIR/webroot"

	local osie_current=$DEPLOYDIR/webroot/misc/osie/current
	local tink_workflow=$DEPLOYDIR/webroot/workflow/
	if [ ! -d "$osie_current" ] && [ ! -d "$tink_workflow" ]; then
		mkdir -p "$osie_current"
		mkdir -p "$tink_workflow"
		pushd "$SCRATCH"

		if [ -z "${TB_OSIE_TAR:-}" ]; then
			curl 'https://tinkerbell-oss.s3.amazonaws.com/osie-uploads/latest.tar.gz' -o ./osie.tar.gz
			tar -zxf osie.tar.gz
		else
			tar -zxf "$TB_OSIE_TAR"
		fi

		if pushd osie*/; then
			if mv workflow-helper.sh workflow-helper-rc "$tink_workflow"; then
				cp -r ./* "$osie_current"
			else
				echo "$ERR failed to move 'workflow-helper.sh' and 'workflow-helper-rc'"
				exit 1
			fi
			popd
		fi
	else
		echo "$INFO found existing osie files, skipping osie setup"
	fi
)

check_container_status() (
	if docker ps | grep -q "$1"; then
		echo "$ERR failed to start container $1"
		exit 1
	fi
)

gen_certs() (
	mkdir -p "$DEPLOYDIR/certs"

	if [ ! -f "$DEPLOYDIR/certs/ca.json" ]; then
		jq \
			'.
			 | .names[0].L = $facility
			' \
			"$DEPLOYDIR/tls/ca.in.json" \
			--arg ip "$TINKERBELL_HOST_IP" \
			--arg facility "$FACILITY" \
			>"$DEPLOYDIR/certs/ca.json"
	fi

	if [ ! -f "$DEPLOYDIR/certs/server-csr.json" ]; then

		jq \
			'.
			| .hosts += [ $ip, "tinkerbell.\($facility).packet.net" ]
			| .names[0].L = $facility
			| .hosts = (.hosts | sorto | unique)
			' \
			"$DEPLOYDIR/tls/server-csr.in.json" \
			--arg ip "$TINKERBELL_HOST_IP" \
			--arg facility "$FACILITY" \
			>"$DEPLOYDIR/certs/server-csr.json"
	fi

	docker build --tag "tinkerbell-certs" "$DEPLOYDIR/tls"
	docker run --rm \
		--volume "$DEPLOYDIR/certs:/certs" \
		--user "$UID:$GID" \
		tinkerbell-certs

	certs_dir="/etc/docker/certs.d/$TINKERBELL_HOST_IP"
	if [ ! -d "$certs_dir" ]; then
		mkdir -p "$certs_dir"
	fi

	# update host to trust registry certificate
	cp "$DEPLOYDIR"/certs/ca.pem "$certs_dir"/tinkerbell.crt

	# copy public key to NGINX for workers
	cp "$DEPLOYDIR"/certs/ca.pem "$DEPLOYDIR/webroot/workflow/ca.pem"
)

generate_certificates() (
	if [ -d "$DEPLOYDIR"/certs ]; then
		echo "$WARN found certs directory"
		if grep -q "\"$TINKERBELL_HOST_IP\"" "$DEPLOYDIR"/tls/server-csr.in.json; then
			echo "$WARN found server entry in TLS"
			echo "$INFO found existing certificates for host $TINKERBELL_HOST_IP, skipping certificate generation"
		else
			gen_certs
		fi
	else
		gen_certs
	fi
)

start_registry() (
	docker-compose -f "$(pwd)"/deploy/docker-compose.yml up --build -d registry
	sleep 5
	check_container_status "registry"

	# push latest worker image to registry
	docker pull quay.io/tinkerbell/tink-worker:latest
	docker tag quay.io/tinkerbell/tink-worker:latest "$TINKERBELL_HOST_IP"/tink-worker:latest
	docker pull fluent/fluent-bit:1.3
	docker tag fluent/fluent-bit:1.3 "$TINKERBELL_HOST_IP"/fluent-bit:1.3
	echo -n "$TINKERBELL_REGISTRY_PASSWORD" | docker login -u="$TINKERBELL_REGISTRY_USERNAME" --password-stdin "$TINKERBELL_HOST_IP"
	docker push "$TINKERBELL_HOST_IP"/tink-worker:latest
	docker push "$TINKERBELL_HOST_IP"/fluent-bit:1.3
)

setup_docker_registry() (
	registry_images=/var/tinkerbell/registry
	if [ ! -d "$registry_images" ]; then
		mkdir -p "$registry_images"
	fi
	if [ -f ~/.docker/config.json ]; then
		if grep -q "$TINKERBELL_HOST_IP" ~/.docker/config.json; then
			echo "$INFO found existing docker auth token for registry $TINKERBELL_HOST_IP, using existing registry"
		else
			start_registry
		fi
	else
		start_registry
	fi
)

start_components() (
	local components=(db cacher hegel tink-server boots tink-cli nginx kibana)
	for comp in "${components[@]}"; do
		docker-compose -f "$(pwd)"/deploy/docker-compose.yml up --build -d "$comp"
		sleep 3
		check_container_status "$comp"
	done
)

command_exists() (
	command -v "$@" >/dev/null 2>&1
)

check_command() (
	if command_exists "$1"; then
		echo "$BLANK Found prerequisite: $1"
		return 0
	else
		echo "$ERR Prerequisite command not installed: $1"
		return 1
	fi
)

check_prerequisites() (
	echo "$INFO verifying prerequisites"
	failed=0
	check_command git || failed=1
	check_command bc || failed=1
	check_command jq || failed=1
	check_command ifup || failed=1
	check_command docker || failed=1
	check_command docker-compose || failed=1

	if [ $failed -eq 1 ]; then
		echo "$ERR Prerequisites not met. Please install the missing commands and re-run $0."
		exit 1
	fi
)

whats_next() (
	echo "$NEXT With the provisioner setup successfully, you can now try executing your first workflow."
	echo "$BLANK Follow the steps described in https://tinkerbell.org/examples/hello-world/ to say 'Hello World!' with a workflow."
)

do_setup() (
	# perform some very rudimentary platform detection
	lsb_dist=$(get_distribution)
	lsb_dist="$(echo "$lsb_dist" | tr '[:upper:]' '[:lower:]')"
	lsb_version=$(get_distro_version)

	echo "$INFO starting tinkerbell stack setup"
	check_prerequisites "$lsb_dist"

	if [ ! -f "$ENV_FILE" ]; then
		echo "$ERR Run './generate-envrc.sh network-interface > \"$ENV_FILE\"' before continuing."
		exit 1
	fi

	# shellcheck disable=SC1090
	source "$ENV_FILE"

	setup_networking "$lsb_dist" "$lsb_version"

	setup_osie
	generate_certificates
	setup_docker_registry
	start_components
	echo ""
	until docker-compose -f "$(pwd)"/deploy/docker-compose.yml ps; do
		sleep 3
		echo ""
	done
	echo "$INFO tinkerbell stack setup completed successfully on $lsb_dist server"
	whats_next
)

# wrapped up in a function so that we have some protection against only getting
# half the file during "curl | sh"
do_setup
