#!/bin/bash

# Check user provided variables
function check_variable() {
    if [ -z "$2" ] ; then
        echo "Error: $1 is not exported"
        exit 1
    fi
}

# Shorthand to print value for a particular variable
function dump_variable() {
    echo "$1=$2"
}

# Check required variables for script
function check_required_input() {
    check_variable TINKERBELL_NETWORK "$TINKERBELL_NETWORK"
    check_variable TINKERBELL_HOST_IP "$TINKERBELL_HOST_IP"
    check_variable TINKERBELL_NGINX_IP "$TINKERBELL_NGINX_IP"
    check_variable TINKERBELL_BROADCAST_IP "$TINKERBELL_BROADCAST_IP"
    check_variable TINKERBELL_REGISTRY_USER "$TINKERBELL_REGISTRY_USER"
    check_variable TINKERBELL_REGISTRY_PASSWORD "$TINKERBELL_REGISTRY_PASSWORD"
    check_variable TINKERBELL_NETMASK "$TINKERBELL_NETMASK"
    check_variable TINKERBELL_PROVISIONER_INTERFACE "$TINKERBELL_PROVISIONER_INTERFACE"

    TINKERBELL_NETMASK_LEN=$(echo $TINKERBELL_NETWORK | grep -Eo "\/[[:digit:]]+" | grep -v "^$" | tr -d "/")

    dump_variable TINKERBELL_NETWORK $TINKERBELL_NETWORK
    dump_variable TINKERBELL_HOST_IP $TINKERBELL_HOST_IP
    dump_variable TINKERBELL_NGINX_IP $TINKERBELL_NGINX_IP
    dump_variable TINKERBELL_BROADCAST_IP $TINKERBELL_BROADCAST_IP
    dump_variable TINKERBELL_REGISTRY_USER $TINKERBELL_REGISTRY_USER
    dump_variable TINKERBELL_NETMASK $TINKERBELL_NETMASK
    dump_variable TINKERBELL_NETMASK_LEN $TINKERBELL_NETMASK_LEN
    dump_variable TINKERBELL_PROVISIONER_INTERFACE $TINKERBELL_PROVISIONER_INTERFACE
}

# Check if interface exists on provisioner
function check_provisioner_interface() {
    if ip addr show | grep $TINKERBELL_PROVISIONER_INTERFACE >> /dev/null; then
        echo "$TINKERBELL_PROVISIONER_INTERFACE found on provisioner."
    else
        echo "Error: Interface $TINKERBELL_PROVISIONER_INTERFACE does not exist on provsioner."
        exit 1
    fi
}

function setup_environemt() { 
    # Below variables will eventually goaway but required for now
    export FACILITY="onprem"
    export ROLLBAR_TOKEN="ignored"
    export ROLLBAR_DISABLE=1

    # export input variables
    export IP_CIDR=$TINKERBELL_NETMASK_LEN
    export BROAD_IP=$TINKERBELL_BROADCAST_IP
    export NETMASK=$TINKERBELL_NETMASK
    if grep "HOST_IP" /etc/environment >> /dev/null; then
        echo "Envs are set already set."
    else
        echo "HOST_IP=$TINKERBELL_HOST_IP" >> /etc/environment
        echo "NGINX_IP=$TINKERBELL_NGINX_IP" >> /etc/environment
        echo "TINKERBELL_REGISTRY_USER=$TINKERBELL_REGISTRY_USER" >> /etc/environment
        echo "TINKERBELL_REGISTRY_PASSWORD=$TINKERBELL_REGISTRY_PASSWORD" >> /etc/environment
        echo "TINKERBELL_GRPC_AUTHORITY=127.0.0.1:42113" >> /etc/environment
        echo "TINKERBELL_CERT_URL=http://127.0.0.1:42114/cert" >> /etc/environment
    fi
}

function setup_network_ubuntu() {
    if [ ! -f /etc/network/interfaces ] ; then
        echo "Error: File /etc/network/interfaces not found"
        exit 1
    fi
    if grep -c $TINKERBELL_HOST_IP /etc/network/interfaces ; then
        echo "Interface $TINKERBELL_PROVISIONER_INTERFACE already configured."
    else 
        # Plumb IP to provisioner interface
        sed -i "/$TINKERBELL_PROVISIONER_INTERFACE inet \(manual\|dhcp\)/c\\iface $TINKERBELL_PROVISIONER_INTERFACE inet static\n    address $TINKERBELL_HOST_IP\n    netmask $TINKERBELL_NETMASK\n    broadcast $TINKERBELL_BROADCAST_IP" /etc/network/interfaces
        # Restart interface
        ifdown  $TINKERBELL_PROVISIONER_INTERFACE
        ifup  $TINKERBELL_PROVISIONER_INTERFACE
        # Add NGINX IP
        if ! sudo ip addr add $TINKERBELL_NGINX_IP/$TINKERBELL_NETMASK_LEN dev $TINKERBELL_PROVISIONER_INTERFACE; then
           echo "Failed to add NGINX IP address to interface $TINKERBELL_PROVISIONER_INTERFACE"
           exit 1
	    fi
        echo "Network configured for interface $TINKERBELL_PROVISIONER_INTERFACE"
    fi
}


function setup_network_centos () {
    if ! systemctl enable NetworkManager; then 
        echo "Failed to enable NetworkManager"
        exit 1
    fi  
    if ! systemctl start NetworkManager; then
        echo "Failed to start NetworkManager"
        exit 1
    fi    
    if ! nmcli con add type ethernet con-name $TINKERBELL_PROVISIONER_INTERFACE ifname $TINKERBELL_PROVISIONER_INTERFACE ip4 $TINKERBELL_HOST_IP/$TINKERBELL_NETMASK_LEN; then
        echo "Failed to add IP address to interface $TINKERBELL_PROVISIONER_INTERFACE"
        exit 1
    fi
    if ! nmcli con up $TINKERBELL_PROVISIONER_INTERFACE; then
        echo "Failed to bring up the interface $TINKERBELL_PROVISIONER_INTERFACE"
        exit 1
    fi

    #ADD NGINX_IP
    if ! ip addr add $TINKERBELL_NGINX_IP/$TINKERBELL_NETMASK_LEN dev $TINKERBELL_PROVISIONER_INTERFACE; then
        echo "Failed to add NGINX IP address to interface $TINKERBELL_PROVISIONER_INTERFACE"
        exit 1
    fi
    echo "Network confihured for interface $TINKERBELL_PROVISIONER_INTERFACE"
}

function validate_network_conf() {
    if ! ip addr show $TINKERBELL_PROVISIONER_INTERFACE | grep $TINKERBELL_HOST_IP >> /dev/null; then
        echo "Interface $TINKERBELL_PROVISIONER_INTERFACE does not have host IP $TINKERBELL_HOST_IP"
        exit 1
    fi
    if ! ip addr show $TINKERBELL_PROVISIONER_INTERFACE | grep $TINKERBELL_NGINX_IP >> /dev/null; then
        echo "Interface $TINKERBELL_PROVISIONER_INTERFACE does not have nginx IP $TINKERBELL_NGINX_IP"
        exit 1
    fi
    echo "Network Configurations are valid"
}


function setup_osie_with_nginx() {
    mkdir -p /etc/tinkerbell/nginx/misc/osie/current
    mkdir -p /etc/tinkerbell/nginx/misc/tinkerbell/workflow/

    pushd /tmp
    curl 'https://tinkerbell-oss.s3.amazonaws.com/osie-uploads/latest.tar.gz' -o osie.tar.gz
    tar -zxvf osie.tar.gz
    if pushd /tmp/osie*/ ; then
        if mv -v workflow-helper.sh workflow-helper-rc /etc/tinkerbell/nginx/misc/tinkerbell/workflow/; then
            cp -rv ./* /etc/tinkerbell/nginx/misc/osie/current/
            rm -v /tmp/latest -rf
        else
            printf "Failed to move \"workflow-helper.sh\" and \"workflow-helper-rc\" files"
            exit 1
        fi
	    popd
    fi
    popd
}

function check_container_error() {
    docker ps | grep $1 >> /dev/null 
    if [ "$?" -eq  "0" ]; then
        echo "No Error" >> /dev/null
    else
      echo "Failed to start $1"
      exit 1
    fi
}

function build_and_setup_certs () {
    if [ ! -d  tls ] ; then
        echo "Error: Directory tls does not exist"
        exit 1
    fi

    grep "$TINKERBELL_HOST_IP" tls/server-csr.in.json >> /dev/null
    if [[ $? -eq 1 ]]
    then            
        sed -i -e "s/localhost\"\,/localhost\"\,\n    \"$TINKERBELL_HOST_IP\"\,/g" tls/server-csr.in.json
    fi

    # build the certificates
    docker-compose up --build -d certs
    sleep 2
    docker ps -a | grep certs | grep "Exited (0)" >> /dev/null
    if [ "$?" -eq "0" ]; then
        sleep 2
    else 
        echo "$LINENO: Certs container was not executed successfully"
        exit 1
    fi

    #Update host to trust registry certificate
    if mkdir -p /etc/docker/certs.d/$TINKERBELL_HOST_IP; then
        cp -v certs/ca.pem /etc/docker/certs.d/$TINKERBELL_HOST_IP/ca.crt
    fi

    mkdir -p /etc/tinkerbell/nginx/misc/tinkerbell/workflow/
    #copy certificate in tinkerbell
    cp -v certs/ca.pem /etc/tinkerbell/nginx/misc/tinkerbell/workflow/ca.pem

}

function build_registry_and_update_worker_image() {
    #Build private registry
    docker-compose up --build -d registry
    sleep 5
    check_container_error "registry"

    #pull the worker image and push into private registry
    docker pull quay.io/tinkerbell/tink-worker:latest
    docker tag quay.io/tinkerbell/tink-worker:latest $TINKERBELL_HOST_IP/tink-worker:latest

    #login to private registry and push the worker image
    docker login -u=$TINKERBELL_REGISTRY_USER -p=$TINKERBELL_REGISTRY_PASSWORD $TINKERBELL_HOST_IP
    docker push $TINKERBELL_HOST_IP/tink-worker:latest
}

function start_docker_stack() {

    docker-compose up --build -d db
    sleep 5
    check_container_error "db"
    
    docker-compose up --build -d tink-server
    sleep 5
    check_container_error "tink-server"

    docker-compose up --build -d nginx
    sleep 5
    check_container_error "nginx"

    docker-compose up --build -d cacher
    sleep 5
    check_container_error "cacher"
    
    docker-compose up --build -d hegel
    sleep 5
    check_container_error "hegel"
    
    docker-compose up --build -d boots
    sleep 5
    check_container_error "boots"

    docker-compose up --build -d kibana
    sleep 2
    check_container_error "kibana"

    docker-compose up --build -d tink-cli
    sleep 2
    check_container_error "tink-cli"
}

function setup_network() {
    # Extract OS
    if [ -f /etc/redhat-release ] && [ $(grep -ci centos /etc/redhat-release) ] ; then
        setup_network_centos
    elif [ -f /etc/os-release ] && [ $(grep -ci ubuntu /etc/os-release) ]; then 
        setup_network_ubuntu
    else
        echo "Error: Unsupported operatings system."
        exit 1
    fi
}

check_required_input
check_provisioner_interface
setup_environemt
setup_network
validate_network_conf
setup_osie_with_nginx
build_and_setup_certs
build_registry_and_update_worker_image
start_docker_stack
