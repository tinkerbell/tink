#!/bin/bash

# abort this script on errors
set -euxo pipefail

whoami

cd /vagrant

ensure_os_packages_exists() (
	declare -a pkgs=()
	for p in "$@"; do
		if ! command_exists "$p"; then
			pkgs+=("$p")
		fi
	done

	if ((${#pkgs[@]} == 0)); then
		return
	fi

	sudo apt-get update
	sudo apt-get install -y "${pkgs[@]}"
)

ensure_docker_exists() (
	if command_exists docker; then
		return
	fi

	# steps from https://docs.docker.com/engine/install/ubuntu/

	ensure_os_packages_exists \
		apt-transport-https \
		ca-certificates \
		gnupg-agent \
		software-properties-common \
		;

	curl -fsSL https://download.docker.com/linux/ubuntu/gpg |
		sudo apt-key add -

	local repo
	repo=$(
		printf "deb [arch=amd64] https://download.docker.com/linux/ubuntu %s stable" \
			"$(lsb_release -cs)"
	)
	sudo add-apt-repository "$repo"

	ensure_os_packages_exists \
		containerd.io \
		docker-ce \
		docker-ce-cli \
		;
)

ensure_docker-compose_exists() (
	if command_exists docker-compose; then
		return
	fi

	# from https://docs.docker.com/compose/install/
	sudo curl -fsSL \
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

configure_vagrant_user() (
	sudo usermod -aG docker vagrant

	echo -n "$TINKERBELL_REGISTRY_PASSWORD" |
		sudo -iu vagrant docker login \
			--username="$TINKERBELL_REGISTRY_USERNAME" \
			--password-stdin "$TINKERBELL_HOST_IP"
)

main() (
	export DEBIAN_FRONTEND=noninteractive

	ensure_os_packages_exists curl jq
	ensure_docker_exists
	ensure_docker-compose_exists

	if [ ! -f ./.env ]; then
		./generate-env.sh eth1 >.env
	fi

	# shellcheck disable=SC1091
	. ./.env

	make_certs_writable

	./setup.sh

	secure_certs

	configure_vagrant_user

)

main
