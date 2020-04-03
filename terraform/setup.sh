#!/bin/bash

function setup_network() {
    declare network_interface=$(grep auto /etc/network/interfaces | tail -1 | cut -d ' ' -f 2)
    echo "This is network interface" $network_interface
    declare bond=$(cat /etc/network/interfaces | tail -1)
    sed -i -e "s/$bond//g" /etc/network/interfaces
    sed -i -e "s/$network_interface inet manual/$network_interface inet static\n    address $HOST_IP\n    netmask $NETMASK\n    broadcast $BROAD_IP/g" /etc/network/interfaces
    ifdown  $network_interface
    ifup  $network_interface

    declare host=$HOST_IP

    declare ip=$(($(echo $host | cut -d "." -f 4 | xargs) + 1))
    declare nginx_ip="$(echo $host | cut -d "." -f 1).$(echo $host | cut -d "." -f 2).$(echo $host | cut -d "." -f 3).$ip"
    echo "This is nginx host" $nginx_ip
    sudo ip addr add $nginx_ip/$IP_CIDR dev $network_interface
    echo "NGINX_IP=$nginx_ip" >> /etc/environment
}

function setup_envs() {
    sudo apt update -y
    sudo apt-get install -y wget ca-certificates

    # export packet variables
    export FACILITY="onprem"
    export PACKET_API_AUTH_TOKEN="dummy_token"
    export PACKET_API_URL=""
    export PACKET_CONSUMER_TOKEN="dummy_token"
    export PACKET_ENV="onprem"
    export PACKET_VERSION="onprem"
    export ROLLBAR_TOKEN="ignored"
    export ROLLBAR_DISABLE=1
}



function install_required_tools() {
    #setup git and git lfs
    sudo apt-get update
    sudo apt-get install -y git
    wget https://github.com/git-lfs/git-lfs/releases/download/v2.9.0/git-lfs-linux-amd64-v2.9.0.tar.gz
    tar -C /usr/local/bin -xzf git-lfs-linux-amd64-v2.9.0.tar.gz
    rm git-lfs-linux-amd64-v2.9.0.tar.gz
    git lfs install

    # Get docker and docker-compose
    curl -L get.docker.com | bash
    curl -L "https://github.com/docker/compose/releases/download/1.24.1/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
    chmod +x /usr/local/bin/docker-compose

    # Setup go
    wget https://dl.google.com/go/go1.13.9.linux-amd64.tar.gz
    tar -C /usr/local -xzf go1.13.9.linux-amd64.tar.gz go/
    rm go1.13.9.linux-amd64.tar.gz

    # set GOPATH
    echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
    echo 'export GOPATH=$GOPATH:$HOME/go' >> ~/.bashrc
    echo 'export PATH=$PATH:$GOPATH' >> ~/.bashrc
    source ~/.bashrc
}

function setup_osie_with_nginx() {
    mkdir -p /packet/nginx/misc/osie/current
    mkdir -p /packet/nginx/misc/tinkerbell/workflow/
    cd /packet/nginx/misc/tinkerbell/workflow/
    wget https://raw.githubusercontent.com/tinkerbell/osie/master/installer/workflow-helper.sh
    wget https://raw.githubusercontent.com/tinkerbell/osie/master/installer/workflow-helper-rc
    chmod +x workflow-helper.sh
    
    cd /tmp
    curl 'https://packet-osie-uploads.s3.amazonaws.com/osie-v19.10.23.00-n=55,c=be58d67,b=master.tar.gz' -o osie.tar.gz
    tar -zxvf osie.tar.gz
    cd /tmp/'osie-v19.10.23.00-n=55,c=be58d67,b=master'
    cp -r -v * /packet/nginx/misc/osie/current/
    rm -v /tmp/'osie-v19.10.23.00-n=55,c=be58d67,b=master' -rf
}


function build_and_setup_certs () {
    grep "$HOST_IP" /etc/network/interfaces
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

    mkdir -p /packet/nginx/misc/tinkerbell/workflow/
    #copy certificate in tinkerbell
    cp certs/ca.pem /packet/nginx/misc/tinkerbell/workflow/ca.pem
}

function build_registry_and_update_worker_image () {
    #pull the worker image and push into private registry
    docker pull quay.io/tinkerbell/tink-worker:master
    docker tag quay.io/tinkerbell/tink-worker:master $host/worker:latest

    # Start the stack
    docker-compose up --build -d registry
    sleep 5

    cd ~/go/src/github.com/packethost/tinkerbell
    #push worker image into it
    docker login -u=$TINKERBELL_REGISTRY_USER -p=$TINKERBELL_REGISTRY_PASS $host
    docker push $host/worker:latest
}

function start_docker_stack() {
    docker-compose up --build -d db
    sleep 5
    docker-compose up --build -d tink-server
    sleep 5
    docker-compose up --build -d nginx
    sleep 5
    docker-compose up --build -d cacher
    sleep 5
    docker-compose up --build -d hegel
    sleep 5
    docker-compose up --build -d boots
    sleep 5
    docker-compose up --build -d kibana
    sleep 2
    docker-compose up --build -d tink-cli
}

install_required_tools;
setup_envs;
setup_network;
setup_osie_with_nginx;


# Give permission to tink binary
chmod +x /usr/local/bin/tink


# get the tinkerbell repo
mkdir -p ~/go/src/github.com/tinkerbell
cd ~/go/src/github.com/tinkerbell
git clone https://github.com/tinkerbell/tink.git
cd ~/go/src/github.com/tinkerbell/tink

build_and_setup_certs;
build_registry_and_update_worker_image;
start_docker_stack;

# Update ip tables
iptables -t nat -I POSTROUTING -s $HOST_IP/$IP_CIDR  -j MASQUERADE
iptables -I FORWARD -d $HOST_IP/$IP_CIDR  -j ACCEPT
iptables -I FORWARD -s $HOST_IP/$IP_CIDR  -j ACCEPT
