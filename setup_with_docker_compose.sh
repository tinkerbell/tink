#!/bin/bash

source input.sh

function install_docker() {
    # Get docker and docker-compose
    curl -L get.docker.com | bash
    curl -L "https://github.com/docker/compose/releases/download/1.24.1/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
    chmod +x /usr/local/bin/docker-compose
}

function setup_environemt() { 
    # Below variables will eventually goaway but required for now
    export FACILITY="onprem"
    export ROLLBAR_TOKEN="ignored"
    export ROLLBAR_DISABLE=1

    # export input variables
    export HOST_IP=$host_ip
    export NGINX_IP=$nginx_ip
    export IP_CIDR=$cidr
    export BROAD_IP=$broad_ip
    export NETMASK=$netmask
    export TINKERBELL_REGISTRY_USER=$private_registry_user
    export TINKERBELL_REGISTRY_PASS=$private_registry_pass
    export TINKERBELL_GRPC_AUTHORITY=127.0.0.1:42113
    export TINKERBELL_CERT_URL=http://127.0.0.1:42114/cert
}

function setup_network () {
    network_interface=$(grep auto /etc/network/interfaces | tail -1 | cut -d ' ' -f 2)
    echo "This is network interface" $network_interface

    sed -i "/$network_interface inet /c\\iface $network_interface inet static\n    address $HOST_IP\n    netmask $NETMASK\n    broadcast $BROAD_IP" /etc/network/interfaces

    ifdown  $network_interface
    ifup  $network_interface

    sudo ip addr add $NGINX_IP/$IP_CIDR dev $network_interface
}

function build_and_setup_certs () {
    sed -i -e "s/localhost\"\,/localhost\"\,\n    \"$HOST_IP\"\,/g" tls/server-csr.in.json

    # build the certificates
    docker-compose up --build -d certs
    sleep 5
    #Update host to trust registry certificate
    mkdir -p /etc/docker/certs.d/$HOST_IP

    cp certs/ca.pem /etc/docker/certs.d/$HOST_IP/ca.crt

    mkdir -p /packet/nginx/misc/tinkerbell/workflow/
    #copy certificate in tinkerbell
    cp certs/ca.pem /packet/nginx/misc/tinkerbell/workflow/ca.pem
}

function build_registry_and_update_worker_image() {
    #Build private registry
    docker-compose up --build -d registry
    sleep 5

    #pull the worker image and push into private registry
    docker pull quay.io/packet/tinkerbell-worker:workflow
    docker tag quay.io/packet/tinkerbell-worker:workflow $HOST_IP/worker:latest

    #login to private registry and push the worker image
    docker login -u=$TINKERBELL_REGISTRY_USER -p=$TINKERBELL_REGISTRY_PASS $HOST_IP
    docker push $HOST_IP/worker:latest
}

function start_docker_stack() {
    docker-compose up --build -d db
    sleep 5
    docker-compose up --build -d tinkerbell
    sleep 5
    docker-compose up --build -d nginx
    sleep 5
    docker-compose up --build -d cserver
    sleep 5
    docker-compose up --build -d hegel
    sleep 5
    docker-compose up --build -d boots
}

#install_docker;
setup_environemt;
setup_network;
build_and_setup_certs;
build_registry_and_update_worker_image;
start_docker_stack;
