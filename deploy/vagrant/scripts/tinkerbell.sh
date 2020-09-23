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

	curl -fsSL https://download.docker.com/linux/$(. /etc/os-release; echo "$ID")/gpg |
		sudo apt-key add -

	local repo
	repo=$(
		printf "deb [arch=amd64] https://download.docker.com/linux/$(. /etc/os-release; echo "$ID") %s stable" \
			"$(lsb_release -cs)"
	)
	sudo add-apt-repository "$repo"

	sudo apt-get update
	sudo apt-get install -y docker-ce docker-ce-cli containerd.io
)

setup_docker_compose() (
	# from https://docs.docker.com/compose/install/
	local DOCKER_COMPOSE_DOWNLOAD_LINK=${DOCKER_COMPOSE_DOWNLOAD_LINK:-"https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)"}  # If variable not set or null, use default.
	sudo curl -C - -SLR --progress-bar \
		"${DOCKER_COMPOSE_DOWNLOAD_LINK}" \
		-o /usr/local/bin/docker-compose

	sudo chmod +x /usr/local/bin/docker-compose
	docker-compose version
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

configure_vagrant_user() (
	sudo usermod -aG docker vagrant

	echo -n "$TINKERBELL_REGISTRY_PASSWORD" |
		sudo -iu vagrant docker login \
			--username="$TINKERBELL_REGISTRY_USERNAME" \
			--password-stdin "$TINKERBELL_HOST_IP"
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

	configure_vagrant_user

)

main
