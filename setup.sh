#!/bin/bash

# stops the execution if a command or pipeline has an error
set -e

# Tinkerbell stack Linux setup script
#
# See https://tinkerbell.org/setup for the installation steps.

# file to hold all environment variables
ENV_FILE=envrc

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

get_distribution() {
	lsb_dist=""
	# Every system that we officially support has /etc/os-release
	if [ -r /etc/os-release ]; then
		lsb_dist="$(. /etc/os-release && echo "$ID")"
	fi
	# Returning an empty string here should be alright since the
	# case statements don't act unless you provide an actual value
	echo "$lsb_dist"
}

is_network_configured() {
	ip addr show $TINKERBELL_PROVISIONER_INTERFACE | grep $TINKERBELL_HOST_IP >>/dev/null && ip addr show $TINKERBELL_PROVISIONER_INTERFACE | grep $TINKERBELL_NGINX_IP >>/dev/null
}

write_iface_config() {
	iface_config="$(
		cat <<EOF | sed 's/^\s\{4\}//g' | sed ':a;N;$!ba;s/\n/\\n/g'
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
	)"
	sed -i "/^auto $TINKERBELL_NETWORK_INTERFACE/,/^\$/c $iface_config" /etc/network/interfaces
}

setup_networking() {
	if ! is_network_configured; then
		case "$1" in
		ubuntu)
			if [ ! -f /etc/network/interfaces ]; then
				echo "$ERR file /etc/network/interfaces not found"
				exit 1
			fi

			if grep -q $TINKERBELL_HOST_IP /etc/network/interfaces; then
				echo "$INFO tinkerbell network interface is already configured"
			else
				# plumb IP and restart to tinkerbell network interface
				if grep -q $TINKERBELL_NETWORK_INTERFACE /etc/network/interfaces; then
					echo "" >>/etc/network/interfaces
					write_iface_config
				else
					echo -e "\nauto $TINKERBELL_NETWORK_INTERFACE\n" >>/etc/network/interfaces
					write_iface_config
				fi
				ip link set $TINKERBELL_NETWORK_INTERFACE nomaster
				ifdown "$TINKERBELL_NETWORK_INTERFACE:0"
				ifdown "$TINKERBELL_NETWORK_INTERFACE:1"
				ifup "$TINKERBELL_NETWORK_INTERFACE:0"
				ifup "$TINKERBELL_NETWORK_INTERFACE:1"
			fi
			;;
		centos)
			if [ -f /etc/sysconfig/network-scripts/ifcfg-$TINKERBELL_NETWORK_INTERFACE ]; then
				sed -i '/^ONBOOT.*no$/s/no/yes/; /^BOOTPROTO.*none$/s/none/static/; /^MASTER/d; /^SLAVE/d' /etc/sysconfig/network-scripts/ifcfg-$TINKERBELL_NETWORK_INTERFACE
			else
				touch /etc/sysconfig/network-scripts/ifcfg-$TINKERBELL_NETWORK_INTERFACE
				HWADDRESS=$(ip addr show $TINKERBELL_NETWORK_INTERFACE | grep ether | awk -F 'ether' '{print $2}' | cut -d" " -f2)
				cat <<EOF >>/etc/sysconfig/network-scripts/ifcfg-$TINKERBELL_NETWORK_INTERFACE
DEVICE=$TINKERBELL_NETWORK_INTERFACE
ONBOOT=yes
HWADDR=$HWADDRESS
BOOTPROTO=static
EOF
			fi

			cat <<EOF >>/etc/sysconfig/network-scripts/ifcfg-$TINKERBELL_NETWORK_INTERFACE
IPADDR0=$TINKERBELL_HOST_IP
NETMASK0=$TINKERBELL_NETMASK
IPADDR1=$TINKERBELL_NGINX_IP
NETMASK1=$TINKERBELL_NETMASK
EOF
			ip link set $TINKERBELL_NETWORK_INTERFACE nomaster
			ifup $TINKERBELL_NETWORK_INTERFACE
			;;
		esac
		if is_network_configured; then
			echo "$INFO tinkerbell network interface configured successfully"
		else
			echo "$ERR tinkerbell network interface configuration failed"
		fi
	else
		echo "$INFO tinkerbell network interface is already configured"
	fi
}

setup_osie() {
	osie_current=/var/tinkerbell/nginx/misc/osie/current
	tink_workflow=/var/tinkerbell/nginx/workflow/
	if [ ! -d "$osie_current" ] && [ ! -d "$tink_workflow" ]; then
		mkdir -p "$osie_current"
		mkdir -p "$tink_workflow"
		pushd /tmp

		if [ -z "${TB_OSIE_TAR:-}" ]; then
			curl 'https://tinkerbell-oss.s3.amazonaws.com/osie-uploads/latest.tar.gz' -o osie.tar.gz
			tar -zxf osie.tar.gz
		else
			tar -zxf "$TB_OSIE_TAR"
		fi

		if pushd /tmp/osie*/; then
			if mv workflow-helper.sh workflow-helper-rc "$tink_workflow"; then
				cp -r ./* "$osie_current"
				rm /tmp/latest -rf
			else
				echo "$ERR failed to move 'workflow-helper.sh' and 'workflow-helper-rc'"
				exit 1
			fi
			popd
		fi
		popd
		rm -f /tmp/osie.tar.gz
	else
		echo "$INFO found existing osie files, skipping osie setup"
	fi
}

check_container_status() {
	docker ps | grep "$1" >>/dev/null
	if [ "$?" -ne "0" ]; then
		echo "$ERR failed to start container $1"
		exit 1
	fi
}

gen_certs() {
	sed -i -e "s/localhost\"\,/localhost\"\,\n    \"$TINKERBELL_HOST_IP\"\,/g" "$deploy"/tls/server-csr.in.json
	docker-compose -f "$deploy"/docker-compose.yml up --build -d certs
	sleep 2
	docker ps -a | grep certs | grep "Exited (0)" >>/dev/null
	if [ "$?" -eq "0" ]; then
		sleep 2
	else
		echo "$ERR failed to generate certificates"
		exit 1
	fi

	certs_dir="/etc/docker/certs.d/$TINKERBELL_HOST_IP"
	if [ ! -d "$certs_dir" ]; then
		mkdir -p "$certs_dir"
	fi

	# update host to trust registry certificate
	cp "$deploy"/certs/ca.pem "$certs_dir"/ca.crt
	# copy public key to NGINX for workers
	cp "$deploy"/certs/ca.pem /var/tinkerbell/nginx/workflow/ca.pem
}

generate_certificates() {
	deploy="$(pwd)"/deploy
	if [ ! -d "$deploy"/tls ]; then
		echo "$ERR directory 'tls' does not exist"
		exit 1
	fi

	if [ -d "$deploy"/certs ]; then
		echo "$WARN found certs directory"
		if grep -q "\"$TINKERBELL_HOST_IP\"" "$deploy"/tls/server-csr.in.json; then
			echo "$WARN found server entry in TLS"
			echo "$INFO found existing certificates for host $TINKERBELL_HOST_IP, skipping certificate generation"
		else
			gen_certs
		fi
	else
		gen_certs
	fi
}

start_registry() {
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
}

setup_docker_registry() {
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
}

start_components() {
	components=(db cacher hegel tink-server boots tink-cli nginx kibana)
	for comp in "${components[@]}"; do
		docker-compose -f "$(pwd)"/deploy/docker-compose.yml up --build -d "$comp"
		sleep 3
		check_container_status "$comp"
	done
}

command_exists() {
	command -v "$@" >/dev/null 2>&1
}

check_command() {
	if command_exists "$1"; then
		echo "$BLANK Found prerequisite: $1"
		return 0
	else
		echo "$ERR Prerequisite command not installed: $1"
		return 1
	fi
}

check_prerequisites() {
	echo "$INFO verifying prerequisites"
	failed=0
	check_command git || failed=1
	check_command ifup || failed=1
	check_command docker || failed=1
	check_command docker-compose || failed=1

	if [ $failed -eq 1 ]; then
		echo "$ERR Prerequisites not met. Please install the missing commands and re-run $0."
		exit 1
	fi
}

whats_next() {
	echo "$NEXT With the provisioner setup successfully, you can now try executing your first workflow."
	echo "$BLANK Follow the steps described in https://tinkerbell.org/examples/hello-world/ to say 'Hello World!' with a workflow."
}

do_setup() {
	# perform some very rudimentary platform detection
	lsb_dist=$(get_distribution)
	lsb_dist="$(echo "$lsb_dist" | tr '[:upper:]' '[:lower:]')"

	echo "$INFO starting tinkerbell stack setup"
	check_prerequisites "$lsb_dist"

	if [ ! -f "$ENV_FILE" ]; then
		echo "$ERR Run './generate-envrc.sh network-interface > \"$ENV_FILE\"' before continuing."
		exit 1
	fi

	source "$ENV_FILE"

	# Run setup for each distro accordingly
	case "$lsb_dist" in
	ubuntu)
		setup_networking "$lsb_dist"
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
		exit 0
		;;
	centos)
		# enable IP forwarding for docker
		echo "net.ipv4.ip_forward=1" >>/etc/sysctl.conf
		setup_networking "$lsb_dist"
		setup_osie
		generate_certificates
		setup_docker_registry
		start_components
		until docker-compose -f "$(pwd)"/deploy/docker-compose.yml ps; do
			sleep 3
			echo ""
		done
		echo "$INFO tinkerbell stack setup completed successfully on $lsb_dist server"
		whats_next
		exit 0
		;;
	*)
		echo
		echo "$ERR unsupported distribution '$lsb_dist'"
		echo
		exit 1
		;;
	esac
	echo "$INFO tinkerbell stack setup failed on $lsb_dist server"
	exit 1
}

# wrapped up in a function so that we have some protection against only getting
# half the file during "curl | sh"
do_setup
