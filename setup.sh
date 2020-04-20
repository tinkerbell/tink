#!/bin/bash

# stops the execution if a command or pipeline has an error
set -e

# Tinkerbell stack Linux setup script
#
# See https://tinkerbell.org/setup for the installation steps.

# file to hold all environment variables 
ENV_FILE=envrc

# color codes
RED="$(tput setaf 1)"
GREEN="$(tput setaf 2)"
YELLOW="$(tput setaf 3)"
RESET="$(tput sgr0)"
INFO="${GREEN}INFO:$RESET"
ERR="${RED}ERROR:$RESET"
WARN="${YELLOW}WARNING:$RESET"

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

list_network_interfaces() {
	echo "Following network interfaces found on the system:"
	ip -o link show | awk -F': ' '{print $2}' | grep '^[e]'
}

get_tinkerbell_network_interface() {
    # read for network interface if TB_INTERFACE is not defined
	if [ -z $TB_INTERFACE ]; then
		read -p 'Which one would you like to use for with Tinkerbell? ' tink_interface	
	else
        tink_interface=$TB_INTERFACE
	fi
	
    ip -o link show | awk -F': ' '{print $2}' | sed 's/[ \t].*//;/^\(lo\|\)$/d' | sed 's/[ \t].*//;/^\(bond0\|\)$/d' | grep "$tink_interface"
    if [ $? -ne 0 ]; then
        echo "$ERR Invalid interface selected. Exiting setup."
        exit 1
    fi	
    echo -e "# Network interface for Tinkerbell \nexport TINKERBELL_NETWORK_INTERFACE=$tink_interface" >> "$ENV_FILE"	
}

get_tinkerbell_ips() {
	# read for tinkerbell network if TB_NETWORK is not defined
	if [ -z $TB_NETWORK ]; then
		read -p 'Select the subnet for Tinkerbell ecosystem: [default 192.168.1.0/29]: ' tink_network
		tink_network=${tink_network:-"192.168.1.0/29"}
	else
		tink_network=$TB_NETWORK
	fi

	# read for tinkerbell IP address if TB_IPADDR is not defined
	if [ -z $TB_IPADDR ]; then
		read -p 'Select the IP address for Tinkerbell [default 192.168.1.1]: ' ip_addr	
		ip_addr=${ip_addr:-"192.168.1.1"}	
	else
		ip_addr=$TB_IPADDR
	fi

	host=$(($(echo $ip_addr | cut -d "." -f 4 | xargs) + 1))
	nginx_ip="$(echo $ip_addr | cut -d "." -f 1).$(echo $ip_addr | cut -d "." -f 2).$(echo $ip_addr | cut -d "." -f 3).$host"

	# calculate network and broadcast based on supplied provide IP range
	if [[ $tink_network =~ ^([0-9\.]+)/([0-9]+)$ ]]; then
		# CIDR notation
		IPADDR=${BASH_REMATCH[1]}
		NETMASKLEN=${BASH_REMATCH[2]}
		zeros=$((32-NETMASKLEN))
		NETMASKNUM=0
		for (( i=0; i<$zeros; i++ )); do
			NETMASKNUM=$(( (NETMASKNUM << 1) ^ 1 ))
		done
		NETMASKNUM=$((NETMASKNUM ^ 0xFFFFFFFF))
		toaddr $NETMASKNUM NETMASK
	else
    	IPADDR=${1:-192.168.1.1}
    	NETMASK=${2:-255.255.255.248}
	fi

	tonum $IPADDR IPADDRNUM
	tonum $NETMASK NETMASKNUM

	# The logic to calculate network and broadcast
	INVNETMASKNUM=$(( 0xFFFFFFFF ^ NETMASKNUM ))
	NETWORKNUM=$(( IPADDRNUM & NETMASKNUM ))
	BROADCASTNUM=$(( INVNETMASKNUM | NETWORKNUM ))

	toaddr $NETWORKNUM NETWORK
	toaddr $BROADCASTNUM BROADCAST

	echo -e "\n# Subnet (IP block) used by Tinkerbell ecosystem \nexport TINKERBELL_NETWORK=$tink_network" >> "$ENV_FILE"	
	echo -e "\n# Host IP is used by provisioner to expose different services such as tink, boots, etc. \nexport TINKERBELL_HOST_IP=$ip_addr" >> "$ENV_FILE"
	echo -e "\n# NGINX IP is used by provisioner to serve files required for iPXE boot \nexport TINKERBELL_NGINX_IP=$nginx_ip" >> "$ENV_FILE"	
	echo -e "\n# Netmask for Tinkerbell network \nexport TINKERBELL_NETMASK=$NETMASK" >> "$ENV_FILE"	
	echo -e "\n# Broadcast IP for Tinkerbell network \nexport TINKERBELL_BROADCAST_IP=$BROADCAST" >> "$ENV_FILE"	
}

tonum() {
    if [[ $1 =~ ([[:digit:]]+)\.([[:digit:]]+)\.([[:digit:]]+)\.([[:digit:]]+) ]]; then
        addr=$(( (${BASH_REMATCH[1]} << 24) + (${BASH_REMATCH[2]} << 16) + (${BASH_REMATCH[3]} << 8) + ${BASH_REMATCH[4]} ))
        eval "$2=\$addr"
    fi
}

toaddr() {
    b1=$(( ($1 & 0xFF000000) >> 24))
    b2=$(( ($1 & 0xFF0000) >> 16))
    b3=$(( ($1 & 0xFF00) >> 8))
    b4=$(( $1 & 0xFF ))
    eval "$2=\$b1.\$b2.\$b3.\$b4"
}

get_registry_credentials() {
	# read for registry username if TB_REGUSER is not defined
	if [ -z $TB_REGUSER ]; then
		read -p 'Create a Docker registry username [default admin]? ' username	
		username=${username:-"admin"}
	else 
		username=$TB_REGUSER
	fi
	password=$(head -c 12 /dev/urandom | sha256sum | cut -d' ' -f1)
	echo -e "\n# We host a private Docker registry on provisioner which is used by different workers" >> "$ENV_FILE"	
	echo -e "# Registry username \nexport TINKERBELL_REGISTRY_USERNAME=$username" >> "$ENV_FILE"	
	echo -e "\n# Registry password \nexport TINKERBELL_REGISTRY_PASSWORD=$password" >> "$ENV_FILE"	
	echo ""
}

generate_envrc() {
	# backup existing environment config if any
	if [ -f "$ENV_FILE" ]; then
    	echo "$INFO found existing $ENV_FILE, moving it to $ENV_FILE.bak"
		mv "$ENV_FILE" "$ENV_FILE".bak
	fi

	list_network_interfaces
	tink_interface=$( get_tinkerbell_network_interface )
	get_tinkerbell_ips
	get_registry_credentials

    # the following envs will eventually goaway but are required for now
	echo -e "\nexport FACILITY=onprem" >> "$ENV_FILE"	
	echo -e "export ROLLBAR_TOKEN=ignored" >> "$ENV_FILE"
	echo -e "export ROLLBAR_DISABLE=1\n" >> "$ENV_FILE"	
}

command_exists() {
	command -v "$@" > /dev/null 2>&1
}

setup_docker() {
	if command_exists docker; then
		echo "$INFO docker already installed, found $(docker -v)"
	else
		echo "$INFO installing docker"
		curl -L get.docker.com | bash   >> /dev/null && echo "$INFO $(docker -v) installed successfully"
	fi

	if command_exists docker-compose; then
		echo "$INFO docker-compose already installed, found $(docker-compose -v)"
	else
		echo "$INFO installing docker-compose"
	 	curl -L "https://github.com/docker/compose/releases/download/1.24.1/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
 		chmod +x /usr/local/bin/docker-compose
	fi
}

is_network_configured() {
    ip addr show $TINKERBELL_PROVISIONER_INTERFACE | grep $TINKERBELL_HOST_IP >> /dev/null && ip addr show $TINKERBELL_PROVISIONER_INTERFACE | grep $TINKERBELL_NGINX_IP >> /dev/null
}

write_iface_config(){
    iface_config="$(cat <<EOF | sed 's/^\s\{4\}//g' | sed ':a;N;$!ba;s/\n/\\n/g'
    iface $TINKERBELL_NETWORK_INTERFACE inet static
        address $TINKERBELL_HOST_IP
        netmask $TINKERBELL_NETMASK
        broadcast $TINKERBELL_BROADCAST_IP
        pre-up sleep 4
EOF
    )"
    sed -i "/^iface $TINKERBELL_NETWORK_INTERFACE/,/^\$/c $iface_config" /etc/network/interfaces
}

setup_networking() {
	if ! is_network_configured ; then
		cidr=$(echo $TINKERBELL_NETWORK | grep -Eo "\/[[:digit:]]+" | grep -v "^$" | tr -d "/")
		case "$1" in 
			ubuntu)
				if [ ! -f /etc/network/interfaces ] ; then
					echo "$ERR file /etc/network/interfaces not found"
					exit 1
				fi

				if grep -q $TINKERBELL_HOST_IP /etc/network/interfaces ; then
				 	echo "$INFO tinkerbell network interface is already configured"
				else 
				 	# plumb IP and restart to tinkerbell network interface
					echo "" >> /etc/network/interfaces
					write_iface_config  
					ifdown "$TINKERBELL_NETWORK_INTERFACE"
					ifup "$TINKERBELL_NETWORK_INTERFACE"					
				fi
				# add NGINX IP
				if ! ip addr add "$TINKERBELL_NGINX_IP/$cidr" dev "$TINKERBELL_NETWORK_INTERFACE"; then
					echo "$ERR failed to add NGINX IP address to network interface - $TINKERBELL_NETWORK_INTERFACE"
					exit 1
				fi
				;;
			centos)
				sed -i '/^ONBOOT.*no$/s/no/yes/; /^BOOTPROTO.*none$/s/none/static/; /^MASTER/d; /^SLAVE/d' /etc/sysconfig/network-scripts/ifcfg-$TINKERBELL_NETWORK_INTERFACE
				cat <<EOF >> /etc/sysconfig/network-scripts/ifcfg-$TINKERBELL_NETWORK_INTERFACE
IPADDR0=$TINKERBELL_HOST_IP
NETMASK0=$TINKERBELL_NETMASK
IPADDR1=$TINKERBELL_NGINX_IP
NETMASK1=$TINKERBELL_NETMASK
EOF
				ip link set $TINKERBELL_NETWORK_INTERFACE nomaster
				systemctl restart network
				;;
		esac
		echo "$INFO tinkerbell network interface configured successfully"
	else 
		echo "$INFO tinkerbell network interface is already configured"
	fi
}

setup_osie() {
	osie_current=/etc/tinkerbell/nginx/misc/osie/current
	tink_workflow=/etc/tinkerbell/nginx/workflow/
	if [ ! -d "$osie_current" ] && [ ! -d "$tink_workflow" ]; then 
		mkdir -p "$osie_current"
    	mkdir -p "$tink_workflow"
    	pushd /tmp
    	curl 'https://tinkerbell-oss.s3.amazonaws.com/osie-uploads/latest.tar.gz' -o osie.tar.gz
    	tar -zxf osie.tar.gz
    	if pushd /tmp/osie*/ ; then
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
	else 
		echo "$INFO found existing osie files, skipping osie setup"
	fi
}

check_container_status() {
    docker ps | grep "$1" >> /dev/null 
    if [ "$?" -ne  "0" ]; then
        echo "$ERR failed to start container $1"
    	exit 1
    fi
}

gen_certs() {
	sed -i -e "s/localhost\"\,/localhost\"\,\n    \"$TINKERBELL_HOST_IP\"\,/g" "$deploy"/tls/server-csr.in.json
	docker-compose -f "$deploy"/docker-compose.yml up --build -d certs
	sleep 2
	docker ps -a | grep certs | grep "Exited (0)" >> /dev/null
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
	cp "$deploy"/certs/ca.pem /etc/tinkerbell/nginx/workflow/ca.pem
}

generate_certificates() {
	deploy="$(pwd)"/deploy
	if [ ! -d  "$deploy"/tls ]; then
        echo "$ERR directory 'tls' does not exist"
        exit 1
    fi

	if [ -d "$deploy"/certs ]; then 
		echo "$WARN found certs directory"
		if grep -q "\"$TINKERBELL_HOST_IP\"" "$deploy"/tls/server-csr.in.json; then
			echo "$WARN found server enty in TLS"
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
	docker login -u="$TINKERBELL_REGISTRY_USERNAME" -p="$TINKERBELL_REGISTRY_PASSWORD" "$TINKERBELL_HOST_IP"
	docker push "$TINKERBELL_HOST_IP"/tink-worker:latest
	docker push "$TINKERBELL_HOST_IP"/fluent-bit:1.3
}

setup_docker_registry() {
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
	for comp in "${components[@]}"
	do
		docker-compose -f "$(pwd)"/deploy/docker-compose.yml up --build -d "$comp"
	    sleep 3
    	check_container_status "$comp"
	done
}

do_setup() {
	echo "$INFO starting tinkerbell stack setup"

	# perform some very rudimentary platform detection
	lsb_dist=$( get_distribution )
	lsb_dist="$(echo "$lsb_dist" | tr '[:upper:]' '[:lower:]')"

	# setup the prerequisites
	generate_envrc
	source $ENV_FILE 
	setup_docker

	# get resources
	# TODO: get from https://raw.githubusercontent.com/tinkerbell/tink/master/deploy.tar.gz
	echo "$INFO getting setup artifacts"	
	wget https://github.com/infracloudio/tink/raw/deploy_stack/deploy.tar.gz
	tar -xf deploy.tar.gz

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
			exit 0
			;;
		centos)
			systemctl start docker
			# enable IP forwarding for docker
			echo "net.ipv4.ip_forward=1" >> /etc/sysctl.conf
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

