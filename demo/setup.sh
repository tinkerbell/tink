#!/bin/bash

source tinkenv

function setup_environment() {
    # Below variables will eventually goaway but required for now
    export FACILITY="onprem"
    export ROLLBAR_TOKEN="ignored"
    export ROLLBAR_DISABLE=1

    # export input variables
    echo "HOST_IP=$host_ip" >> /etc/environment
    echo "NGINX_IP=$nginx_ip" >> /etc/environment
    export IP_CIDR=$cidr
    export BROAD_IP=$broad_ip
    export NETMASK=$netmask
    echo "TINKERBELL_REGISTRY_USER=$private_registry_user" >> /etc/environment
    echo "TINKERBELL_REGISTRY_PASS=$private_registry_pass" >> /etc/environment
    echo "TINKERBELL_GRPC_AUTHORITY=127.0.0.1:42113" >> /etc/environment
    echo "TINKERBELL_CERT_URL=http://127.0.0.1:42114/cert" >> /etc/environment
}

function setup_network () {
    network_interface=$(grep auto /etc/network/interfaces | tail -1 | cut -d ' ' -f 2)
    echo "This is the network interface" $network_interface

    grep "$HOST_IP" /etc/network/interfaces
    if [[ $? -eq 1 ]]
    then
        sed -i "/$network_interface inet \(manual\|dhcp\)/c\\iface $network_interface inet static\n    address $HOST_IP\n    netmask $NETMASK\n    broadcast $BROAD_IP" /etc/network/interfaces
    fi
    ifdown  $network_interface
    ifup  $network_interface

    sudo ip addr add $NGINX_IP/$IP_CIDR dev $network_interface
}

function setup_osie_with_nginx() {
    mkdir -p /packet/nginx/misc/osie/current
    mkdir -p /packet/nginx/workflow/
    cd /packet/nginx/workflow/
    wget https://raw.githubusercontent.com/tinkerbell/osie/master/installer/workflow-helper.sh
    wget https://raw.githubusercontent.com/tinkerbell/osie/master/installer/workflow-helper-rc
    chmod +x workflow-helper.sh

    cd /tmp
    curl 'https://packet-osie-uploads.s3.amazonaws.com/osie-v19.10.23.00-n=55,c=be58d67,b=master.tar.gz' -o osie.tar.gz
    tar -zxvf osie.tar.gz
    cd /tmp/'osie-v19.10.23.00-n=55,c=be58d67,b=master'
    cp -r grub /packet/nginx/misc/osie/current/
    cp modloop-x86_64 /packet/nginx/misc/osie/current/
    cp initramfs-x86_64 /packet/nginx/misc/osie/current/
    cp vmlinuz-x86_64 /packet/nginx/misc/osie/current/
    rm /tmp/'osie-v19.10.23.00-n=55,c=be58d67,b=master' -rf

    cd /packet/nginx/misc/osie/current
    curl 'https://packet-osie-uploads.s3.amazonaws.com/ubuntu_18_04.tar.gz'
    tar -zxvf ubuntu_18_04.tar.gz
    rm ubuntu_18_04.tar.gz
}
function build_and_setup_certs () {
#     sudo apt update -y
     sudo apt-get install -y wget ca-certificates

    cd ~/go/src/github.com/tinkerbell/tink
    grep "$HOST_IP" tls/server-csr.in.json
    if [[ $? -eq 1 ]]
    then
        sed -i -e "s/localhost\"\,/localhost\"\,\n    \"$HOST_IP\"\,/g" tls/server-csr.in.json
    fi
    # build the certificates
    docker-compose up --build -d certs
    sleep 5
    #Update host to trust registry certificate
    mkdir -p /etc/docker/certs.d/$HOST_IP
    cp certs/ca.pem /etc/docker/certs.d/$HOST_IP/ca.crt

    #copy certificate in tinkerbell
    cp certs/ca.pem /packet/nginx/workflow/ca.pem
}

function build_registry_and_update_worker_image() {
    #Build private registry
    docker-compose up --build -d registry
    sleep 5

    #pull the worker image and push into private registry
    docker pull quay.io/tinkerbell/tink-worker:latest
    docker tag quay.io/tinkerbell/tink-worker:latest $HOST_IP/tink-worker:latest

    #login to private registry and push the worker image
    docker login -u=$TINKERBELL_REGISTRY_USER -p=$TINKERBELL_REGISTRY_PASS $HOST_IP
    docker push $HOST_IP/tink-worker:latest
}

function start_docker_stack() {

    docker-compose up --build -d db
    sleep 5
    docker-compose up --build -d cacher
    sleep 5
    docker-compose up --build -d tink-server
    sleep 5
    docker-compose up --build -d nginx
    sleep 5
    docker-compose up --build -d hegel
    sleep 5
    docker-compose up --build -d boots
    sleep 5
    docker-compose up --build -d kibana
    sleep 2
    docker-compose up --build -d tink-cli
}

function update_iptables() {
    iptables -t nat -I POSTROUTING -s $HOST_IP/$IP_CIDR  -j MASQUERADE
    iptables -I FORWARD -d $HOST_IP/$IP_CIDR  -j ACCEPT
    iptables -I FORWARD -s $HOST_IP/$IP_CIDR  -j ACCEPT
}

setup_environment;
setup_network;
setup_osie_with_nginx;
build_and_setup_certs;
build_registry_and_update_worker_image;
start_docker_stack;
update_iptables;