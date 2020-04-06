#!/usr/bin/env bash

function init() {
	# Color codes
	RED='\033[0;31m'
	GREEN='\033[0;32m'
	YELLOW='\033[0;33m'
	BLUE='\033[0;34m'
	MAGENTA='\033[0;35m'
	CYAN='\033[0;36m'
	WHITE='\033[0;37m'
	BYELLOW='\033[0;33;5;7m'
	NC='\033[0m' # No Color

	set -o errexit -o pipefail -o xtrace
}

function rainbow() {
	echo -e "$RED:RED"
	echo -e "$GREEN:GREEN"
	echo -e "$YELLOW:YELLOW"
	echo -e "$BLUE:BLUE"
	echo -e "$MAGENTA:MAGENTA"
	echo -e "$CYAN:CYAN"
	echo -e "$WHITE:WHITE"
	echo -e "$BYELLOW:BYELLOW"
	echo -e "$NC:NC"
}

# syntax: phone_home 1.2.3.4 '{"this": "data"}'
function phone_home() {
	local tink_host=$1
	shift

	puttink "${tink_host}" phone-home "$@"
}

# syntax: problem 1.2.3.4 '{"problem":"something is wrong"}'
function problem() {
	local tink_host=$1
	shift

	puttink "${tink_host}" problem "$@"
}

# syntax: fail 1.2.3.4 "reason"
function fail() {
	local tink_host=$1
	shift

	puttink "${tink_host}" phone-home '{"type":"failure", "reason":"'"$1"'"}'
}

# syntax: tink POST 1.2.3.4 phone-home '{"this": "data"}'
function tink() {
	local method=$1 tink_host=$2 endpoint=$3 post_data=$4

	curl \
		-f \
		-vvvvv \
		-X "${method}" \
		-H "Content-Type: application/json" \
		-d "${post_data}" \
		"${tink_host}/${endpoint}"
}

# syntax: puttink 1.2.3.4 phone-home '{"this": "data"}'
function puttink() {
	local tink_host=$1 endpoint=$2 post_data=$3

	tink "PUT" "${tink_host}" "${endpoint}" "${post_data}"
}

# syntax: posttink 1.2.3.4 phone-home '{"this": "data"}'
function posttink() {
	local tink_host=$1 endpoint=$2 post_data=$3

	tink "POST" "${tink_host}" "${endpoint}" "${post_data}"
}

function dns_resolvers() {
	declare -ga resolvers

	# shellcheck disable=SC2207
	resolvers=($(awk '/^nameserver/ {print $2}' /etc/resolv.conf))
	if [ ${#resolvers[@]} -eq 0 ]; then
		resolvers=("147.75.207.207" "147.75.207.208")
	fi
}

function dns_redhat() {
	local filename=$1
	shift

	for ((i = 0; i <= $((${#resolvers[*]} - 1)); i++)); do
		echo "DNS$((i + 1))=${resolvers[i]}" >>"${filename}"
	done
}

function dns_resolvconf() {
	local filename=$1
	shift

	printf 'nameserver %s\n' "${resolvers[@]}" >"${filename}"
}

function filter_bad_devs() {
	# 7 = loopback devices (can go away)
	# 251, 253, 259 = virtio disk (for qemu tests)
	# others = SCSI block disks
	grep -vE '^(7|8|6[5-9]|7[01]|12[89]|13[0-5]|25[139]):'
}

# args: device,...
# exits if args are not block or loop devices
function assert_block_or_loop_devs() {
	local baddevs
	if baddevs=$(lsblk -dnro 'MAJ:MIN' "$@" | filter_bad_devs) && [[ -n $baddevs ]]; then
		echo "$0: All devices may only be block or loop devices" >&2
		echo "$baddevs" >&2
		exit 1
	fi
}

# args: device,...
# exits if args are not of same device type
function assert_same_type_devs() {
	# shellcheck disable=SC2207
	local majors=($(lsblk -dnro 'MAJ:MIN' "$@" | awk -F: '{print $1}' | sort -u))
	if [[ ${majors[*]} =~ 7 ]] && ((${#majors[*]} > 1)); then
		echo "$0: loop back devices can't be mixed with physical devices"
		exit 1
	fi
}

# syntax: is_/loop_dev device,...
# returns 0 if true, 1 if false
function is_loop_dev() {
	loopdev=1
	if [[ $(lsblk -dnro 'MAJ:MIN' "$@") == 7:* ]]; then
		loopdev=0
	fi
	return $loopdev
}

# syntax: is_uefi,...
# returns 0 if true, 1 if false
function is_uefi() {
	[[ -d /sys/firmware/efi ]]
}

efi_device() {
	local efi_path="$1"
	[ -n "$efi_path" ] && shift || efi_path="/boot/efi"
	findmnt -n --target "$efi_path" "$@"
}

find_uuid_boot_id() {
	efibootmgr -v | grep "$1" | sed 's/^Boot\([0-9a-f]\{4\}\).*/\1/gI;t;d'
}

# syntax: name key
# expects metadata in stdin
# safely sets $name=$metadata[$key]
# accepts a default value as third param
function set_from_metadata() {
	local var=$1 key=$2
	local val
	val=$(jq -r "select(.$key != null) | .$key")
	if [[ -z $val ]]; then
		echo "$key is missing, empty or null" >&2
		if [[ -z $3 ]]; then
			return 1
		else
			echo "using default value $val for $key" >&2
			val=$3
		fi
	fi

	declare -g "$var=$val"
}

# syntax: argvalue name switch
# returns 0 if argvalue is not empty, 1 otherwise after printing to stderr
# the message printed to stderr will be "$0: No $name was provided, $switch is required"
function check_required_arg() {
	arg=$1
	name=$2
	switch=$3
	if [[ -n $arg ]]; then
		return 0
	fi
	echo "$0: No $name was provided, $switch is required." >&2
	return 1
}

# usage: assert_all_args_consumed OPTIND $@
# asserts that the caller did not pass in any extra arguments that are not
# handled by getopts
function assert_all_args_consumed() {
	local index=$1
	shift
	if ((index != $# + 1)); then
		echo "unexpected positional argument: OPTIND:$index args:$*" >&2
		exit 1
	fi
}

# usage: assert_num_disks hwtype num_disks
function assert_num_disks() {
	local hwtype=$1 ndisks=$2
	local -A type2disks=(
		[baremetal_0]=1
		[baremetal_1]=2
		[baremetal_1e]=1
		[baremetal_2]=6
		[baremetal_2a]=1
		[baremetal_2a2]=1
		[baremetal_2a4]=1
		[baremetal_2a5]=1
		[baremetal_2a6]=1
		[baremetal_3]=3
		[baremetal_hua]=1
		[baremetal_s]=14
	)

	((ndisks >= type2disks[hwtype]))
}

# usage: assert_storage_size hwtype blockdev...
function assert_storage_size() {
	# TODO: remove when https://github.com/shellcheck/issues/1213 is closed
	# shellcheck disable=SC2034
	local hwtype=$1
	shift
	local gig=$((1024 * 1024 * 1024))
	local -A type2storage=(
		[baremetal_0]=$((80 * gig))
		[baremetal_1]=$((2 * 120 * gig))
		[baremetal_1e]=$((240 * gig))
		[baremetal_2]=$((6 * 480 * gig))
		[baremetal_2a]=$((340 * gig))
		[baremetal_2a2]=1
		[baremetal_2a4]=1
		[baremetal_2a5]=1
		[baremetal_2a6]=1
		[baremetal_3]=$(((2 * 120 + 1600) * gig))
		[baremetal_hua]=1
		[baremetal_s]=$(((12 * 2048 + 2 * 480) * gig))
	)

	local got=0 sz=0
	for disk; do
		sz=$(blockdev --getsize64 "$disk")
		got=$((got + sz))
	done
	((got >= type2storage[hwtype]))
}

# usage: should_stream $image_url
# returns 0 if image size is unknown or larger than available space in destination
# returns 1 otherwise
function should_stream() {
	local image=$1
	local dest=$2

	available=$(BLOCKSIZE=1 df --output=avail "$dest" | grep -v Avail)
	img_size=$(curl -s -I "$image" | tr -d '\r' | awk 'tolower($0) ~ /content-length/ { print $2 }')
	max_size=$((available - (1024 * 1024 * 1024))) # be safe and allow 1G of leeway

	# img_size == 0 is if server can't stat the file, for example some
	# backend is dynamically generating the file for whatever reason
	if ((img_size == 0)) || ((img_size >= max_size)); then
		return 0
	else
		return 1
	fi
}

# rand31s returns a stream of non-negative random 31-bit integers as uint32s
rand31s() {
	od -An -td4 -w4 </dev/urandom | grep -v '^\s*-' | sed 's|^\s\+||'
}

# rand31 returns a non-negative random 31-bit integer as an uint32
rand31() {
	rand31s | head -n1
}

# rand63s returns a stream of non-negative random 63-bit integers as uint64s
rand63s() {
	od -An -td8 -w8 </dev/urandom | grep -v '^\s*-' | sed 's|^\s\+||'
}

# rand63 returns a non-negative random 63-bit integer as an uint64
rand63() {
	rand63s | head -n1
}

function get_disk_block_size_and_count() {
	local blockdevout
	blockdevout="$(blockdev --getpbsz --getsize64 "$1" | tr '\n' ' ')"
	# shellcheck disable=SC2086
	set -- $blockdevout
	local bs=$1 size=$2
	echo "$bs $((size / bs))"
}

# wipe_check_prep writes a random pattern randomly throughout the disk. It
# returns a string representation of its actions which is meant to be passed
# verbatim to wipe_check_verify
function wipe_check_prep() {
	local dev=$1

	local ret
	ret=$(get_disk_block_size_and_count "$bd")
	# shellcheck disable=SC2086
	set -- $ret
	local bs=$(($1 + 0)) blocks=$(($2 + 0)) sha0
	local lastblock=$((blocks - 1))
	sha0=$(dd if=/dev/zero bs=$bs count=1 status=none | sha1sum | awk '{print $1}')

	local -A hindexes=()
	while ((${#hindexes[@]} < 10)); do
		read -r rand
		index=$((rand % lastblock))
		hindexes[$index]=true
	done < <(rand63s)

	local indexes
	# shellcheck disable=SC2207
	indexes=($(echo "${!hindexes[@]}" | tr ' ' '\n' | sort -n))

	for i in "${indexes[@]}"; do
		base64 -w0 </dev/urandom | head -c $bs | dd of="$dev" seek="$i" bs=$bs status=none conv=notrunc
	done
	echo "$dev $bs $sha0 ${indexes[*]}"
}

# wipe_check_verify verifies that a disk was successfully wiped
function wipe_check_verify() {
	local dev=$1
	local bs=$2
	local sha0=$3
	shift 3
	for index; do
		if ! dd if="$dev" skip="$index" bs="$bs" count=1 status=none |
			sha1sum --status -c <(echo "$sha0  -"); then
			return 1
		fi
	done
}

# slow_wipe is the slowest and least preffered method to wipe a disk, it is
# meant as a fallback for when both blkdiscard and sg_unmap fail.
function slow_wipe() {
	local bd=$1

	# Clear MD superblocks
	# doesn't matter if not part of md devices, mdadm just ignores it
	# "$bd"* expands to full disk and any partitions, yay
	echo "$bd: clear any MD device info"
	mdadm --zero-superblock "$bd"*

	echo "$bd: wipefs"
	wipefs -a "$bd"

	echo "$bd: zap all partition information"
	# sgdisk will complain if corrupt partition table but still zaps everything
	sgdisk -Z "$bd" || :

	echo "$bd: create a single full disk partition"
	sgdisk -o -n 1:0:0 "$bd"

	echo "$bd: re-zapping all partition information"
	sgdisk -Z "$bd"

	local ret
	ret=$(get_disk_block_size_and_count "$bd")
	# shellcheck disable=SC2086
	set -- $ret
	local bs=$(($1 + 0)) blocks=$(($2 + 0))

	echo "$bd: slow wipe using dd"
	# ensure main partition table is wiped
	dd if=/dev/zero of="$bd" bs=$bs count=4096 conv=notrunc status=none
	# now the backup
	dd if=/dev/zero of="$bd" bs=$bs count=4096 seek=$((blocks - 4096)) conv=notrunc status=none

	local slice slices=64
	local chunksize=$((blocks / slices))
	local count=$((256 * 1024 * 1024 / bs)) # delete 256MB of data per chunk
	for slice in $(seq $slices); do
		echo "$slice/$slices"
		dd if=/dev/zero of="$bd" bs=$bs count=$count seek=$(((slice - 1) * chunksize)) status=none
	done
}

# wipe will try it's hardest to wipe a disk of data, successively trying
# `blkdiscard`, `sg_unmap`, and `slow_wipe`. It verifies that random data in
# random locations were zeroed by `blkdiscard` or `sg_unmap`.
function wipe() {
	local disk=$1

	local wipe_check_state
	wipe_check_state=$(wipe_check_prep "$disk")
	# shellcheck disable=SC2086
	blkdiscard "$disk" && wipe_check_verify $wipe_check_state && return
	echo "$disk: blkdiscard failed, trying sg_unmap"

	local last_lba
	last_lba=$(sg_readcap "$disk" |
		awk '/Last logical block address/ {split($4, b, "="); print b[2]}' || :)

	# shellcheck disable=SC2086
	sg_unmap --lba=0 --num="$last_lba" "$disk" && wipe_check_verify $wipe_check_state && return

	echo "$disk: sg_unmap failed, wiping using separate tools"
	slow_wipe "$disk"
}

# fast_wipe will 'clear' data in the fastest possible manner
# This is not a secure wipe and should not be used when data security is required
function fast_wipe() {
	local disk=$1
	{
		blkdiscard "$disk" || :          # I think can sometimes fails
		sgdisk -Z "$disk"                # should never fail
		mdadm --zero-superblock "$disk"* # doesn't fail even if non found
	} >&2
}

# marvell_reset uses mvcli to reset the raid card to JBODs
# usage: megaraid_reset disk...
function marvell_reset() {
	# dmidecode prints error messages on stdout!!!!
	systemmfg=$(dmidecode -s system-manufacturer | head -1)
	echo "Marvell hardware raid device is present on system mfg: $systemmfg"

	echo "Marvell-MVCLI - Deleting all VDs"
	vds=$(mvcli info -o vd | awk '/id:/ {print $2}')
	for vd in $vds; do
		echo "Marvell-MVCLI - Deleting VD id:$vd"
		echo y | mvcli delete -f -o vd -i "$vd"
	done
}

# perc_reset uses perccli to reset the raid card to JBODs
# usage: perc_reset disk...
function perc_reset() {
	# dmidecode prints error messages on stdout!!!!
	systemmfg=$(dmidecode -s system-manufacturer | head -1)
	percmodel=$(perccli64 show all | grep PERC | awk '{print $2}')
	echo "Dell PERC hardware raid device is present on system mfg: $systemmfg"

	#Query controller for drive state smart alert info
	#NOTE: disks in JBOD do not appear as Online or GOOD. Show all slot info for err state
	if perccli64 /call/eall/sall show all | grep "S.M.A.R.T alert flagged by drive" | grep No >/dev/null; then
		echo "PERCCLI - Controller drive state - OK"
	else
		echo "PERCCLI - Controller drive state has problem with SMART data alert! FAIL"
		exit 1
	fi

	#Check/set personality
	if perccli64 /c0 show personality | grep "Current Personality" | grep "HBA-Mode" >/dev/null; then
		echo "PERCCLI - Controller in HBA-Mode - OK"
	elif [[ $percmodel == 'PERCH710PMini' || $percmodel == 'PERCH740PMini' ]]; then
		echo "PERCCLI - Skipping set HBA-Mode. This $percmodel does not support HBA mode"
	else
		echo "PERCCLI - Setting personality to HBA-Mode"
		perccli64 /c0 set personality=HBA
	fi

	#Check/delete all VDs!
	if perccli64 /c0 /vall show | grep "No VDs" >/dev/null; then
		echo "PERCCLI - No VDs configured - OK"
	else
		echo "PERCCLI - Deleting all VDs"
		#This also resets all other configs as well per Dell
		perccli64 /c0 /vall delete force
	fi

	#Check for jbod and enable if needed
	if perccli64 /c0 show jbod | grep "JBOD      ON" >/dev/null; then
		echo "PERCCLI - JBOD is on - OK"
	elif [[ $percmodel == 'PERCH710PMini' || $percmodel == 'PERCH740PMini' ]]; then
		echo "PERCCLI - Skipping set JBOD since $percmodel does not support it"
	else
		echo "PERCCLI - Enable JBOD"
		perccli64 /c0 set jbod=on force
	fi

	if [[ $percmodel == 'PERCH710PMini' || $percmodel == 'PERCH740PMini' ]]; then
		percdisk=$(perccli64 /call/eall/sall show all | grep "[0-9]:[0-9]" | awk '{print $1}' | head -1)
		echo "PERCCLI - Creating RAID0 HW RAID on $percmodel"
		perccli64 /c0 add vd r0 name=RAID0 drives="$percdisk"
		sleep 5
	fi
}

# megaraid_reset uses MegaCli64 to reset the raid card to JBODs
# usage: megaraid_reset disk...
function megaraid_reset() {
	# dmidecode prints error messages on stdout!!!!
	systemmfg=$(dmidecode -s system-manufacturer | head -1)
	echo "LSI hardware raid device is present on system mfg: $systemmfg"

	enc=$(MegaCli64 -EncInfo -a0 | awk '/Device ID/ {print $4}')
	slots=$(MegaCli64 -PDList -a0 | awk '/^Slot Number/ {print $3}')

	echo "LSI-MegaCLI - Disabling battery warning at boot"
	MegaCli64 -AdpSetProp BatWarnDsbl 1 -a0

	echo "LSI-MegaCLI - Marking physical devices on adapter 0 as 'Good'"
	for slot in $slots; do
		info=$(MegaCli64 -PDInfo -PhysDrv "[$enc:$slot]" -a0 | sed -n '/^Firmware state: / s|Firmware state: ||p')
		! [[ $info =~ bad ]] && continue
		MegaCli64 -PDMakeGood -PhysDrv "[$enc:$slot]" -Force -a0
	done

	echo "LSI-MegaCLI - Clearing controller of any foreign configs"
	MegaCli64 -CfgForeign -Clear -a0

	echo "LSI-MegaCLI - Clearing controller config to defaults"
	MegaCli64 -CfgClr -a0

	echo "LSI-MegaCLI - Deleting all LDs"
	MegaCli64 -CfgLdDel -LALL -a0

	echo "LSI-MegaCLI - Configuring controller as JBOD"
	MegaCli64 -AdpSetProp -EnableJBOD -0 -a0
	MegaCli64 -AdpSetProp -EnableJBOD -1 -a0
	for slot in $slots; do
		info=$(MegaCli64 -PDInfo -PhysDrv "[$enc:$slot]" -a0 | sed -n '/^Firmware state: / s|Firmware state: ||p')
		[[ $info =~ JBOD ]] && continue
		MegaCli64 -PDMakeJBOD -PhysDrv "[$enc:$slot]" -a0
	done

	if ! [[ $systemmfg =~ Dell ]]; then
		MegaCli64 -AdpSetProp -EnableJBOD -0 -a0
		echo "Creating pseudo JBOD config on the controller"
		echo "LSI-MegaCLI - Creating JBOD with single disk raid0 arrays"
		MegaCli64 -CfgEachDskRaid0 WT RA Direct NoCachedBadBBU -a0
	fi

	sleep 5
	udevadm settle
}

# detect_os detects the target os by first calling `lsb_release` in the rootdir
# via `chroot`, falling back to using the patched lsb_release bash script
# embedded in osie.
# usage: detect_os $rootdir
# returns 2 strings: os version
function detect_os() {
	local rootdir os version
	rootdir=$1

	os=$(chroot "$rootdir" lsb_release -si | sed 's/ //g' || :)
	version=$(chroot "$rootdir" lsb_release -sr || :)
	[[ -n $os ]] && [[ -n $version ]] && echo "$os $version" && return

	os=$(ROOTDIR=$rootdir ./packet_lsb_release -si)
	version=$(ROOTDIR=$rootdir ./packet_lsb_release -sr)
	echo "$os $version"
}

# use xmlstarlet to find the value of a specific matched element (or attribute)
function xml_ev() {
	local _xml="${1}" _match="${2}" _value="${3}"

	(echo "${_xml}" | xmlstarlet sel -t -m "${_match}" -v "${_value}") || echo ""
}

# use xmlstarlet to select a specific XML element and all elements contained within
function xml_elem() {
	local _xml="${1}" _match="${2}"

	echo "${_xml}" | xmlstarlet sel -t -c "${_match}"
}

# convert a bash associative array to json
function bash_aa_to_json() {
	local _json=""
	eval "local -A _f_array=""${1#*=}"

	# shellcheck disable=SC2154
	for k in "${!_f_array[@]}"; do
		if [ "${_json}" = "" ]; then
			_json="\"${k}\": \"${_f_array[k]}\""
		else
			_json="${_json}, \"${k}\": \"${_f_array[k]}\""
		fi
	done

	_json="{ ${_json} }"
	echo -n "${_json}"
}

function set_root_pw() {
	# TODO
	# FIXME: make sure we don't log pwhash whenever osie logging to kibana happens
	# TODO
	echo -e "${GREEN}#### Setting rootpw${NC}"
	sed -i "s|^root:[^:]*|root:$1|" "$2"
	grep '^root' "$2"
}

function vmlinuz_version() {
	local kernel=$1

	set +o pipefail
	type=$(file -b "$kernel")
	case "$type" in
	*MS-DOS*)
		echo 'kernel is type MS-DOS' >&2 # huawei devs mostly
		strings <"$kernel" | sed -n 's|^Linux version \(\S\+\).*|\1|p'
		;;
	*gzip*)
		echo 'kernel is type gzip' >&2 # 2a
		gunzip <"$kernel" | strings | sed -n 's|^Linux version \(\S\+\).*|\1|p'
		;;
	*bzImage*)
		echo 'kernel is type bzImage' >&2 # x86_64
		# shellcheck disable=SC2001
		echo "$type" | sed 's|.*, version \(\S\+\) .*|\1|'
		;;
	esac
	set -o pipefail
}

function gethost() {
	python3 -c "import urllib3;host=urllib3.util.parse_url('$1').host;assert host;print(host)" || :
}

function is_reachable() {
	local host
	host=$(gethost "$1" | sed 's|^\[\(.*\)]$||')

	if [[ $host =~ ^[.*]$ ]]; then
		echo "host is an ipv6 address, thats not supported" >&2 && exit 1
	fi

	ping -c1 -W1 "$host" &>/dev/null
}

function reacquire_dhcp() {
	dhclient -1 "$1"
}

function ensure_reachable() {
	local url=$1

	echo -e "${YELLOW}###### Checking connectivity to \"$url\"...${NC}"
	if ! is_reachable "$url"; then
		echo -e "${YELLOW}###### Failed${NC}"
		echo -e "${YELLOW}###### Reacquiring dhcp for publicly routable ip...${NC}"
		reacquire_dhcp "$(ip_choose_if)"
		echo -e "${YELLOW}###### OK${NC}"
	fi
	echo -e "${YELLOW}###### Verifying connectivity to custom url host...${NC}"
	is_reachable "$url"
	echo -e "${YELLOW}###### OK${NC}"
}

# determine the default interface to use if ip=dhcp is set
# uses "PACKET_BOOTDEV_MAC" kopt value if it exists
# if none, will use the first "eth" interface that has a carrier link
# falls back to the first "eth" interface alphabetically
# keep sync'ed with installer/alpine/init-*64
# shellcheck disable=SC2019
# shellcheck disable=SC2018
ip_choose_if() {
	local mac
	mac=$(echo "${PACKET_BOOTDEV_MAC:-}" | tr 'A-Z' 'a-z')
	if [ -n "$mac" ]; then
		for x in /sys/class/net/eth*; do
			[ -e "$x" ] && grep -q "$mac" "$x/address" && echo "${x##*/}" && return
		done
	fi

	for x in /sys/class/net/eth*; do
		[ -e "$x" ] && ip link set "${x##*/}" up
	done

	sleep 1

	for x in /sys/class/net/eth*; do
		[ -e "$x" ] && grep -q 1 "$x" && echo "${x##*/}" && return
	done

	for x in /sys/class/net/eth*; do
		[ -e "$x" ] && echo "${x##*/}" && return
	done
}

add_post_install_service() {
	local target="$1"
	install -Dm700 target-files/bin/packet-post-install.sh "$target/bin/packet-post-install.sh"

	cp -v target-files/services/packet-post-install.service "$target/etc/systemd/system/packet-post-install.service"
	ln -s /etc/systemd/system/packet-post-install.service "$target/etc/systemd/system/multi-user.target.wants/packet-post-install.service"
}
