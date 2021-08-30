#!/usr/bin/env bash

# stops the execution if a command or pipeline has an error
set -eu

# Tinkerbell stack Linux setup script
#
# See https://tinkerbell.org/setup for the installation steps.

# file to hold all environment variables
ENV_FILE=.env

SCRATCH=$(mktemp -d -t tmp.XXXXXXXXXX)
readonly SCRATCH
function finish() (
	rm -rf "$SCRATCH"
)
trap finish EXIT

DEPLOYDIR=$(pwd)/deploy
readonly DEPLOYDIR
readonly STATEDIR=$DEPLOYDIR/state

if command -v tput >/dev/null && tput setaf 1 >/dev/null 2>&1; then
	# color codes
	RED="$(tput setaf 1)"
	GREEN="$(tput setaf 2)"
	YELLOW="$(tput setaf 3)"
	RESET="$(tput sgr0)"
fi

INFO="${GREEN:-}INFO:${RESET:-}"
ERR="${RED:-}ERROR:${RESET:-}"
WARN="${YELLOW:-}WARNING:${RESET:-}"
BLANK="      "
NEXT="${GREEN:-}NEXT:${RESET:-}"

get_distribution() (
	local lsb_dist=""
	# Every system that we officially support has /etc/os-release
	if [ -r /etc/os-release ]; then
		# shellcheck disable=SC1091
		lsb_dist="$(. /etc/os-release && echo "$ID")"
	fi
	# Returning an empty string here should be alright since the
	# case statements don't act unless you provide an actual value
	echo "$lsb_dist" | tr '[:upper:]' '[:lower:]'
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

identify_network_strategy() (
	local distro=$1
	local version=$2

	case "$distro" in
	ubuntu)
		if jq -n --exit-status '$distro_version >= 17.10' --argjson distro_version "$version" >/dev/null 2>&1; then
			echo "setup_networking_netplan"
		else
			echo "setup_networking_ubuntu_legacy"
		fi
		;;
	centos)
		echo "setup_networking_centos"
		;;
	*)
		echo "setup_networking_manually"
		;;
	esac
)

setup_networking() (
	local distro=$1
	local version=$2

	setup_network_forwarding

	if is_network_configured; then
		echo "$INFO tinkerbell network interface is already configured"
		return 0
	fi

	local strategy
	strategy=$(identify_network_strategy "$distro" "$version")

	"${strategy}" "$distro" "$version" # execute the strategy

	if is_network_configured; then
		echo "$INFO tinkerbell network interface configured successfully"
	else
		echo "$ERR tinkerbell network interface configuration failed"
	fi
)

setup_networking_manually() (
	local distro=$1
	local version=$2

	echo "$ERR this setup script cannot configure $distro ($version)"
	echo "$BLANK please read this script's source and configure it manually."
	exit 1
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
    address $TINKERBELL_HOST_IP/$TINKERBELL_CIDR
    pre-up sleep 4

auto $TINKERBELL_NETWORK_INTERFACE:1
iface $TINKERBELL_NETWORK_INTERFACE:1 inet static
    address $TINKERBELL_NGINX_IP/$TINKERBELL_CIDR
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
PREFIX0=$TINKERBELL_CIDR
IPADDR1=$TINKERBELL_NGINX_IP
PREFIX1=$TINKERBELL_CIDR
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
	mkdir -p "$STATEDIR/webroot"

	local osie_current=$STATEDIR/webroot/misc/osie/current
	local tink_workflow=$STATEDIR/webroot/workflow/
	if [ ! -d "$osie_current" ] || [ ! -d "$tink_workflow" ]; then
		mkdir -p "$osie_current"
		mkdir -p "$tink_workflow"
		pushd "$SCRATCH"

		if [ -z "${TB_OSIE_TAR:-}" ]; then
			curl -fsSL 'https://tinkerbell-oss.s3.amazonaws.com/osie-uploads/latest.tar.gz' -o ./osie.tar.gz
			tar -zxf osie.tar.gz
		else
			if [ ! -f "$TB_OSIE_TAR" ]; then
				echo "$ERR osie tar not found in the given location $TB_OSIE_TAR"
				exit 1
			fi
			echo "$INFO extracting osie tar"
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
	local container_name="$1"
	local container_id
	container_id=$(docker-compose -f "$DEPLOYDIR/docker-compose.yml" ps -q "$container_name")

	local start_moment
	local current_status
	start_moment=$(docker inspect "${container_id}" --format '{{ .State.StartedAt }}')
	current_status=$(docker inspect "${container_id}" --format '{{ .State.Health.Status }}')

	case "$current_status" in
	starting)
		: # move on to the events check
		;;
	healthy)
		return 0
		;;
	unhealthy)
		echo "$ERR $container_name is already running but not healthy. status: $current_status"
		exit 1
		;;
	*)
		echo "$ERR $container_name is already running but its state is a mystery. status: $current_status"
		exit 1
		;;
	esac

	local status
	read -r status < <(docker events \
		--since "$start_moment" \
		--filter "container=$container_id" \
		--filter "event=health_status" \
		--format '{{.Status}}')

	if [ "$status" != "health_status: healthy" ]; then
		echo "$ERR $container_name is not healthy. status: $status"
		exit 1
	fi
)

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
	if ! cmp --quiet "$STATEDIR"/certs/ca.pem "$STATEDIR/webroot/workflow/ca.pem"; then
		cp "$STATEDIR"/certs/ca.pem "$STATEDIR/webroot/workflow/ca.pem"
	fi

	# update host to trust registry certificate
	if ! cmp --quiet "$STATEDIR/certs/ca.pem" "$certs_dir/tinkerbell.crt"; then
		if [ ! -d "$certs_dir/tinkerbell.crt" ]; then
			# The user will be told to create the directory
			# in the next block, if copying the certs there
			# fails.
			mkdir -p "$certs_dir" || true >/dev/null 2>&1
		fi
		if ! cp "$STATEDIR/certs/ca.pem" "$certs_dir/tinkerbell.crt"; then
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

docker_login() (
	echo -n "$TINKERBELL_REGISTRY_PASSWORD" | docker login -u="$TINKERBELL_REGISTRY_USERNAME" --password-stdin "$TINKERBELL_HOST_IP"
)

# This function takes an image specified as first parameter and it tags and
# push it using the second one. useful to proxy images from a repository to
# another.
docker_mirror_image() (
	local from=$1
	local to=$2

	docker pull "$from"
	docker tag "$from" "$to"
	docker push "$to"
)

start_registry() (
	docker-compose -f "$DEPLOYDIR/docker-compose.yml" up --build -d registry
	check_container_status "registry"
)

# This function supposes that the registry is up and running.
# It configures with the required dependencies.
bootstrap_docker_registry() (
	docker_login

	docker_mirror_image "quay.io/tinkerbell/tink-worker:latest" "${TINKERBELL_HOST_IP}/tink-worker:latest"
)

setup_docker_registry() (
	local registry_images="$STATEDIR/registry"
	if [ ! -d "$registry_images" ]; then
		mkdir -p "$registry_images"
	fi
	start_registry
	bootstrap_docker_registry
)

start_components() (
	local components=(db hegel tink-server boots tink-cli nginx)
	for comp in "${components[@]}"; do
		docker-compose -f "$DEPLOYDIR/docker-compose.yml" up --build -d "$comp"
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
	distro=$1
	version=$2

	echo "$INFO verifying prerequisites for $distro ($version)"
	failed=0
	check_command docker || failed=1
	check_command docker-compose || failed=1
	check_command ip || failed=1
	check_command jq || failed=1

	strategy=$(identify_network_strategy "$distro" "$version")
	case "$strategy" in
	"setup_networking_netplan")
		check_command netplan || failed=1
		;;
	"setup_networking_ubuntu_legacy")
		check_command ifdown || failed=1
		check_command ifup || failed=1
		;;
	"setup_networking_centos")
		check_command ifdown || failed=1
		check_command ifup || failed=1
		;;
	"setup_networking_manually")
		echo "$WARN this script cannot automatically configure your network."
		;;
	*)
		echo "$ERR bug: unhandled network strategy: $strategy"
		exit 1
		;;
	esac

	if [ $failed -eq 1 ]; then
		echo "$ERR Prerequisites not met. Please install the missing commands and re-run $0."
		exit 1
	fi
)

whats_next() (
	echo "$NEXT  1. Enter /vagrant/deploy and run: source ../.env; docker-compose up -d"
	echo "$BLANK 2. Try executing your first workflow."
	echo "$BLANK    Follow the steps described in https://tinkerbell.org/examples/hello-world/ to say 'Hello World!' with a workflow."
)

setup_nat() (
	iptables -A FORWARD -i eth1 -o eth0 -j ACCEPT
	iptables -A FORWARD -i eth0 -o eth1 -m state --state ESTABLISHED,RELATED -j ACCEPT
	iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE
)

do_setup() (
	# perform some very rudimentary platform detection
	lsb_dist=$(get_distribution)
	lsb_version=$(get_distro_version)

	echo "$INFO starting tinkerbell stack setup"
	check_prerequisites "$lsb_dist" "$lsb_version"

	if [ ! -f "$ENV_FILE" ]; then
		echo "$ERR Run './generate-env.sh network-interface > \"$ENV_FILE\"' before continuing."
		exit 1
	fi

	# shellcheck disable=SC1090
	source "$ENV_FILE"

	setup_networking "$lsb_dist" "$lsb_version"
	setup_nat
	setup_osie
	generate_certificates
	setup_docker_registry

	echo "$INFO tinkerbell stack setup completed successfully on $lsb_dist server"
	whats_next
)

# wrapped up in a function so that we have some protection against only getting
# half the file during "curl | sh"
do_setup
