#!/bin/bash

# abort this script on errors
set -euxo pipefail

whoami

cd /vagrant

setup_docker() (
	# steps from https://docs.docker.com/engine/install/ubuntu/
	sudo apt-get install -y \
		apt-transport-https \
		ca-certificates \
		curl \
		gnupg-agent \
		software-properties-common

	curl -fsSL https://download.docker.com/linux/ubuntu/gpg |
		sudo apt-key add -

	local repo
	repo=$(
		printf "deb [arch=amd64] https://download.docker.com/linux/ubuntu %s stable" \
			"$(lsb_release -cs)"
	)
	sudo add-apt-repository "$repo"

	sudo apt-get update
	sudo apt-get install -y docker-ce docker-ce-cli containerd.io

	sudo usermod -aG docker "$USER"

	newgrp
)

setup_docker_compose() (
	# from https://docs.docker.com/compose/install/
	sudo curl -L \
		"https://github.com/docker/compose/releases/download/1.26.0/docker-compose-$(uname -s)-$(uname -m)" \
		-o /usr/local/bin/docker-compose

	sudo chmod +x /usr/local/bin/docker-compose
)

make_certs_writable() (
	local certdir="/etc/docker/certs.d/$TINKERBELL_HOST_IP"
	sudo mkdir -p "$certdir"
	sudo chown -R "$USER" "$certdir"
)

secure_certs() (
	local certdir="/etc/docker/certs.d/$TINKERBELL_HOST_IP"
	sudo chown "root" "$certdir"
)

command_exists() (
	command -v "$@" >/dev/null 2>&1
)

mirror_hello_world() (
	# push the hello-world workflow action image
	docker pull hello-world
	docker tag hello-world "$TINKERBELL_HOST_IP/hello-world"
	docker push "$TINKERBELL_HOST_IP/hello-world"
)

main() (
	export DEBIAN_FRONTEND=noninteractive

	apt-get update

	if ! command_exists docker; then
		setup_docker
	fi

	if ! command_exists docker-compose; then
		setup_docker_compose
	fi

	if ! command_exists jq; then
		sudo apt-get install -y jq
	fi

	if [ ! -f ./envrc ]; then
		./generate-envrc.sh eth1 >envrc
	fi

	# shellcheck disable=SC1091
	. ./envrc

	make_certs_writable

	./setup.sh

	secure_certs

	mirror_hello_world

	cd deploy
	docker-compose up -d
)

main
