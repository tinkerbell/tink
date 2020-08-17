#!/usr/bin/env bash

YUM="yum"
APT="apt"
PIP3="pip3"
YUM_CONFIG_MGR="yum-config-manager"
WHICH_YUM=$(command -v $YUM)
WHICH_APT=$(command -v $APT)
YUM_INSTALL="$YUM install"
APT_INSTALL="$APT install"
PIP3_INSTALL="$PIP3 install"
declare -a YUM_LIST=("https://download.docker.com/linux/centos/7/x86_64/stable/Packages/containerd.io-1.2.6-3.3.el7.x86_64.rpm"
	"docker-ce"
	"docker-ce-cli"
	"epel-release"
	"python3")
declare -a APT_LIST=("docker"
	"docker-compose")

add_yum_repo() (
	$YUM_CONFIG_MGR --add-repo https://download.docker.com/linux/centos/docker-ce.repo
)

update_yum() (
	$YUM_INSTALL -y yum-utils
	add_yum_repo
)

update_apt() (
	$APT update
	DEBIAN_FRONTEND=noninteractive $APT --yes --force-yes -o Dpkg::Options::="--force-confdef" -o Dpkg::Options::="--force-confold" upgrade
)

restart_docker_service() (
	service docker restart
)

install_yum_packages() (
	$YUM_INSTALL "${YUM_LIST[@]}" -y
)

install_pip3_packages() (
	$PIP3_INSTALL docker-compose
)

install_apt_packages() (
	$APT_INSTALL "${APT_LIST[@]}" -y
)

main() (
	if [[ -n $WHICH_YUM ]]; then
		update_yum
		install_yum_packages
		install_pip3_packages
		restart_docker_service
	elif [[ -n $WHICH_APT ]]; then
		update_apt
		install_apt_packages
		restart_docker_service
	else
		echo "Unknown platform. Error while installing the required packages"
		exit 1
	fi
)

main
