#!/bin/bash

source tinkenv

function initial_install() {
  # install Go
  wget https://dl.google.com/go/go1.13.9.linux-amd64.tar.gz
  tar -C /usr/local -xzf go1.13.9.linux-amd64.tar.gz go/
  rm go1.13.9.linux-amd64.tar.gz

  # set GOPATH
  echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
  echo 'export GOPATH=$GOPATH:$HOME/go' >> ~/.bashrc
  echo 'export PATH=$PATH:$GOPATH' >> ~/.bashrc
  source ~/.bashrc

  # install Docker and Docker Compose
  curl -L get.docker.com | bash
  curl -L "https://github.com/docker/compose/releases/download/1.24.1/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
  chmod +x /usr/local/bin/docker-compose
}

function setup_network() {
    network_interface=$(grep auto /etc/network/interfaces | tail -1 | cut -d ' ' -f 2)
    echo "This is the network interface" $network_interface

    grep "$HOST_IP" /etc/network/interfaces
    if [[ $? -eq 1 ]]
    then
        declare bond=$(cat /etc/network/interfaces | tail -1)
        sed -i -e "s/$bond//g" /etc/network/interfaces
        sed -i "/$network_interface inet \(manual\|dhcp\)/c\\iface $network_interface inet static\n    address $HOST_IP\n    netmask $NETMASK\n    broadcast $BROAD_IP" /etc/network/interfaces
    fi
    ifdown  $network_interface
    ifup  $network_interface

    sudo ip addr add $NGINX_IP/$IP_CIDR dev $network_interface
}

function setup_osie_with_nginx() {
    mkdir -p /etc/tinkerbell/nginx/misc/osie/current
    mkdir -p /etc/tinkerbell/nginx/workflow/
    cd /etc/tinkerbell/nginx/workflow/
    wget https://raw.githubusercontent.com/tinkerbell/osie/master/installer/workflow-helper.sh
    wget https://raw.githubusercontent.com/tinkerbell/osie/master/installer/workflow-helper-rc
    chmod +x workflow-helper.sh

    cd /tmp
    curl 'https://tinkerbell-oss.s3.amazonaws.com/osie-uploads/latest.tar.gz' -o osie.tar.gz
    mkdir osie-latest
    tar -zxvf osie.tar.gz -C osie-latest --strip-components 1
    cd /tmp/'osie-latest'
    cp -r grub /etc/tinkerbell/nginx/misc/osie/current/
    cp modloop-x86_64 /etc/tinkerbell/nginx/misc/osie/current/
    cp initramfs-x86_64 /etc/tinkerbell/nginx/misc/osie/current/
    cp vmlinuz-x86_64 /etc/tinkerbell/nginx/misc/osie/current/
    rm /tmp/'osie-latest' -rf

    cd /etc/tinkerbell/nginx/misc/osie/current
    curl 'https://packet-osie-uploads.s3.amazonaws.com/ubuntu_18_04.tar.gz' -o ubuntu_18_04.tar.gz
    tar -zxvf ubuntu_18_04.tar.gz
    rm ubuntu_18_04.tar.gz
}
function build_and_setup_certs () {
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
    # update host to trust registry certificate
    mkdir -p /etc/docker/certs.d/$HOST_IP
    cp certs/ca.pem /etc/docker/certs.d/$HOST_IP/ca.crt

    # copy certificate in tinkerbell
    cp certs/ca.pem /etc/tinkerbell/nginx/workflow/ca.pem
}

function build_registry_and_update_worker_image() {
    # build private registry
    docker-compose up --build -d registry
    sleep 5

    # pull the worker image and push into private registry
    docker pull quay.io/tinkerbell/tink-worker:latest
    docker tag quay.io/tinkerbell/tink-worker:latest $HOST_IP/tink-worker:latest

    # login to private registry and push the worker image
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

initial_install;
setup_network;
setup_osie_with_nginx;
build_and_setup_certs;
build_registry_and_update_worker_image;
start_docker_stack;
update_iptables;