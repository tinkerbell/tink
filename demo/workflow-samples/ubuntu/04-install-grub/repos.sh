#!/bin/bash

source functions.sh && init

#defaults

USAGE="Usage: $0 -t /mnt/target
Required Arguments:
	-a arch     System architecture {aarch64|x86_64}
	-M metadata File containing instance metadata
	-t target   Target mount point to write repos to

Options:
	-f facility Facility to use to reach artifacts server
	-m url      Address to embed into OS for package repository (advanced usage, default http://mirror.\$facility.packet.net)
	-h          This help message
	-v          Turn on verbose messages for debugging

Description: This script will configure the package repositories based upon LSB properties located on a target mount point.
"
while getopts "a:M:t:f:m:hv" OPTION; do
	case $OPTION in
	a) arch=$OPTARG ;;
	M) metadata=$OPTARG ;;
	t) export TARGET="$OPTARG" ;;
	f) facility="$OPTARG" ;;
	m) mirror="$OPTARG" ;;
	h) echo "$USAGE" && exit 0 ;;
	v) set -x ;;
	*) echo "$USAGE" && exit 1 ;;
	esac
done

check_required_arg "$arch" "arch" "-a"
check_required_arg "$metadata" 'metadata file' '-M'
check_required_arg "$TARGET" 'target mount point' '-t'
assert_all_args_consumed "$OPTIND" "$@"

# if $mirror is not empty then the user specifically passed in the mirror
# location, we should not trample it
mirror=${mirror:-http://mirror.$facility.packet.net}

# Make sure target provided is mounted
if grep -qs "$TARGET" /proc/mounts; then
	echo "Target is mounted... good."
else
	echo "Error: Target $TARGET is not mounted"
	exit 1
fi

os_ver=$(detect_os "$TARGET")
# shellcheck disable=SC2086
set -- $os_ver
DOS=$1
DVER=$2
echo "#### Detected OS on mounted target $TARGET"
echo "OS: $DOS  ARCH: $arch VER: $DVER"

# Match detected OS to known OS config
do_centos() {
	echo "Configuring repos for CentOS"

	cat <<-EOF_yum_repo >"$TARGET/etc/yum.repos.d/packet.repo"
		[packet-base]
		name=CentOS-\$releasever - Base
		baseurl=${mirror}/centos/\$releasever/os/\$basearch/
		gpgcheck=1
		enabled=1
		gpgkey=file:///etc/pki/rpm-gpg/RPM-GPG-KEY-CentOS-7
		priority=10
		
		#released updates
		[packet-updates]
		name=CentOS-\$releasever - Updates
		baseurl=${mirror}/centos/\$releasever/updates/\$basearch/
		gpgcheck=1
		enabled=1
		gpgkey=file:///etc/pki/rpm-gpg/RPM-GPG-KEY-CentOS-7
		priority=10
		
		[packet-extras]
		name=CentOS-\$releasever - Extras
		baseurl=${mirror}/centos/\$releasever/extras/\$basearch/
		gpgcheck=0
		enabled=1
		gpgkey=file:///etc/pki/rpm-gpg/RPM-GPG-KEY-CentOS-7
		priority=10
	EOF_yum_repo

	sed -i '/distroverpkg=centos-release/a exclude=microcode_ctl' "$TARGET/etc/yum.conf"
}

do_ubuntu() {
	case "$DVER" in
	'14.04') do_ubuntu_14_04 ;;
	'16.04') "do_ubuntu_16_04_$arch" ;;
	'17.10') "do_ubuntu_17_10_$arch" ;;
	'18.04') "do_ubuntu_18_04_$arch" ;;
	'19.04') "do_ubuntu_19_04_$arch" ;;
	*) do_unknown ;;
	esac
}

do_ubuntu_14_04() {
	echo "Configuring repos for Ubuntu $DVER"

	cat <<-EOF_ub_repo >"$TARGET/etc/apt/sources.list"
		deb [arch=amd64] ${mirror}/ubuntu trusty main universe
		deb [arch=amd64] ${mirror}/ubuntu trusty-updates main universe
		deb [arch=amd64] ${mirror}/ubuntu trusty-security main universe
	EOF_ub_repo
}

do_ubuntu_16_04_x86_64() {
	echo "Configuring repos for Ubuntu $DVER"
	cat <<-EOF_ub_repo >"$TARGET/etc/apt/sources.list"
		deb http://archive.ubuntu.com/ubuntu/ xenial main universe
		deb http://archive.ubuntu.com/ubuntu/ xenial-updates main universe
		deb http://archive.ubuntu.com/ubuntu/ xenial-security main universe
	EOF_ub_repo
}

do_ubuntu_16_04_aarch64() {
	echo "Configuring repos for Ubuntu $DVER for $arch"
	echo 'Acquire::ForceIPv4 "true";' >"$TARGET/etc/apt/apt.conf.d/99force-ipv4"
	cat <<-EOF_ub_repo >"$TARGET/etc/apt/sources.list"
		deb http://ports.ubuntu.com/ubuntu-ports xenial main multiverse universe
		deb http://ports.ubuntu.com/ubuntu-ports xenial-backports main multiverse universe
		deb http://ports.ubuntu.com/ubuntu-ports xenial-security main multiverse universe
		deb http://ports.ubuntu.com/ubuntu-ports xenial-updates main multiverse universe
	EOF_ub_repo
}

do_ubuntu_17_10_x86_64() {
	echo "Configuring repos for Ubuntu $DVER"
	cat <<-EOF_ub_repo >"$TARGET/etc/apt/sources.list"
		deb http://archive.ubuntu.com/ubuntu/ artful main universe
		deb http://archive.ubuntu.com/ubuntu/ artful-updates main universe
		deb http://archive.ubuntu.com/ubuntu/ artful-security main universe
	EOF_ub_repo

	cat <<-EOF_ub_motd >"$TARGET/etc/motd"
		=====================================================================
		-                               NOTICE                              -
		-  Ubuntu 17.10 is/will be EOL and Packet.net will discontinue support  -
		-  for it soon. Please consider upgrading or moving to LTS. For     -
		-  more information please see https://bit.ly/2glHLU8                -
		=====================================================================
	EOF_ub_motd
}

do_ubuntu_17_10_aarch64() {
	echo "Configuring repos for Ubuntu $DVER for $arch"
	echo 'Acquire::ForceIPv4 "true";' >"$TARGET/etc/apt/apt.conf.d/99force-ipv4"
	cat <<-EOF_ub_repo >"$TARGET/etc/apt/sources.list"
		deb http://ports.ubuntu.com/ubuntu-ports artful main multiverse universe
		deb http://ports.ubuntu.com/ubuntu-ports artful-backports main multiverse universe
		deb http://ports.ubuntu.com/ubuntu-ports artful-security main multiverse universe
		deb http://ports.ubuntu.com/ubuntu-ports artful-updates main multiverse universe
	EOF_ub_repo

	cat <<-EOF_ub_motd >"$TARGET/etc/motd"
		=====================================================================
		-                               NOTICE                              -
		-  Ubuntu 17.10 is/will be EOL and Packet.net will discontinue support  -
		-  for it soon. Please consider upgrading or moving to LTS. For     -
		-  more information please see https://bit.ly/2glHLU8                -
		=====================================================================
	EOF_ub_motd
}

do_ubuntu_18_04_x86_64() {
	echo "Configuring repos for Ubuntu $DVER"
	cat <<-EOF_ub_repo >"$TARGET/etc/apt/sources.list"
		deb http://archive.ubuntu.com/ubuntu bionic main restricted
		deb http://archive.ubuntu.com/ubuntu bionic-updates main restricted
		deb http://archive.ubuntu.com/ubuntu bionic universe
		deb http://archive.ubuntu.com/ubuntu bionic-updates universe
		deb http://archive.ubuntu.com/ubuntu bionic multiverse
		deb http://archive.ubuntu.com/ubuntu bionic-updates multiverse
		deb http://archive.ubuntu.com/ubuntu bionic-backports main restricted universe multiverse
		deb http://security.ubuntu.com/ubuntu bionic-security main restricted
		deb http://security.ubuntu.com/ubuntu bionic-security universe
		deb http://security.ubuntu.com/ubuntu bionic-security multiverse
	EOF_ub_repo
}

do_ubuntu_18_04_aarch64() {
	echo "Configuring repos for Ubuntu $DVER for $arch"
	echo 'Acquire::ForceIPv4 "true";' >"$TARGET/etc/apt/apt.conf.d/99force-ipv4"
	cat <<-EOF_ub_repo >"$TARGET/etc/apt/sources.list"
		deb http://ports.ubuntu.com/ubuntu-ports bionic main multiverse universe
		deb http://ports.ubuntu.com/ubuntu-ports bionic-backports main multiverse universe
		deb http://ports.ubuntu.com/ubuntu-ports bionic-security main multiverse universe
		deb http://ports.ubuntu.com/ubuntu-ports bionic-updates main multiverse universe
	EOF_ub_repo
}

do_ubuntu_19_04_x86_64() {
	echo "Configuring repos for Ubuntu $DVER"
	cat <<-EOF_ub_repo >"$TARGET/etc/apt/sources.list"
		deb http://archive.ubuntu.com/ubuntu disco main restricted
		deb http://archive.ubuntu.com/ubuntu disco-updates main restricted
		deb http://archive.ubuntu.com/ubuntu disco universe
		deb http://archive.ubuntu.com/ubuntu disco-updates universe
		deb http://archive.ubuntu.com/ubuntu disco multiverse
		deb http://archive.ubuntu.com/ubuntu disco-updates multiverse
		deb http://archive.ubuntu.com/ubuntu disco-backports main restricted universe multiverse
		deb http://security.ubuntu.com/ubuntu disco-security main restricted
		deb http://security.ubuntu.com/ubuntu disco-security universe
		deb http://security.ubuntu.com/ubuntu disco-security multiverse
	EOF_ub_repo
}

do_ubuntu_19_04_aarch64() {
	echo "Configuring repos for Ubuntu $DVER for $arch"
	echo 'Acquire::ForceIPv4 "true";' >"$TARGET/etc/apt/apt.conf.d/99force-ipv4"
	cat <<-EOF_ub_repo >"$TARGET/etc/apt/sources.list"
		deb http://ports.ubuntu.com/ubuntu-ports disco main multiverse universe
		deb http://ports.ubuntu.com/ubuntu-ports disco-backports main multiverse universe
		deb http://ports.ubuntu.com/ubuntu-ports disco-security main multiverse universe
		deb http://ports.ubuntu.com/ubuntu-ports disco-updates main multiverse universe
	EOF_ub_repo
}

do_debian() {
	case "$DVER" in
	'8.8') do_debian_8 ;;
	9.*) do_debian_9 ;;
	10.*) do_debian_10 ;;
	*) do_unknown ;;
	esac
}

do_debian_8() {
	if [ "$arch" = "aarch64" ]; then
		echo "Nothing to do for Debian arm64"
		exit 0
	fi

	echo "Configuring repos for Debian $DVER"

	cat <<-EOF_deb_repo >"$TARGET/etc/apt/sources.list"
		deb [arch=amd64] http://security.debian.org jessie/updates main non-free contrib
		deb [arch=amd64] ${mirror}/debian jessie-backports main
		deb [arch=amd64] ${mirror}/debian jessie main non-free contrib
		deb [arch=amd64] ${mirror}/debian jessie-updates main non-free contrib
	EOF_deb_repo
}

do_debian_9() {
	if [ "$arch" = "aarch64" ]; then
		echo "Nothing to do for Debian arm64"
		exit 0
	fi

	echo "Configuring repos for Debian $DVER"

	cat <<-EOF_deb_repo >"$TARGET/etc/apt/sources.list"
		deb [arch=amd64] http://security.debian.org stretch/updates main non-free contrib
		deb [arch=amd64] ${mirror}/debian stretch-backports main
		deb [arch=amd64] ${mirror}/debian stretch main non-free contrib
		deb [arch=amd64] ${mirror}/debian stretch-updates main non-free contrib
	EOF_deb_repo
}

do_debian_10() {
	if [ "$arch" = "aarch64" ]; then
		echo "Nothing to do for Debian arm64"
		exit 0
	fi

	echo "Configuring repos for Debian $DVER"

	cat <<-EOF_deb_repo >"$TARGET/etc/apt/sources.list"
		deb [arch=amd64] http://security.debian.org buster/updates main non-free contrib
		deb [arch=amd64] ${mirror}/debian buster-backports main
		deb [arch=amd64] ${mirror}/debian buster main non-free contrib
		deb [arch=amd64] ${mirror}/debian buster-updates main non-free contrib
	EOF_deb_repo
}

do_scientificcernslc() {
	echo "Nothing to do for SciLinux"
}

do_redhatenterpriseserver() {
	echo "Nothing to do for RedHatEnterpriseServer"
}

do_unknown() {
	echo "Warning: Detected OS $DOS not matched! It's up to the image repo now."
}

# TODO: maybe break off in into repo-$DOS-$arch.sh, error if ! test -x?
case "$DOS" in
'CentOS' | 'RedHatEnterpriseServer' | 'Debian' | 'ScientificCERNSLC' | 'Ubuntu') "do_${DOS,,}" ;;
*) do_unknown ;;
esac
exit 0
