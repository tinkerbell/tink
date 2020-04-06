#!/bin/bash

source functions.sh && init
set -o nounset

#Description: This script will configure the disk partition layout, filesystem, fstab and software
#raid config (if applicable) based upon Packet metadata.

fstab=/tmp/fstab.tmpl

config=$1
target=$2
preserve_data=$3
deprovision_fast=$4

check_required_arg "$config" 'cpr config' 'config as first param'
check_required_arg "$target" 'target directory' 'target as second param'
check_required_arg "$preserve_data" 'should data volumes be preserved' 'preserve_data as third param'
check_required_arg "$deprovision_fast" 'should disk wiping be skipped' 'deprovision_fast as fourth param'

stormeta() {
	jq -r ".$1" "$config"
}

msg() {
	echo -e "${@}" "${NC}" >&2
}

add_to_fstab() {
	local UUID=${1} MOUNTPOINT=${2} FSFORMAT=${3} options=errors=remount-ro fsck=2
	if grep -qs "${UUID}" $fstab; then
		msg "WARNING: Not adding ${UUID} to fstab again (it's already there!)"
		return
	fi

	if [[ ${FSFORMAT} == swap ]]; then
		fsck=0
		options=none
	fi
	[[ ${MOUNTPOINT} == '/' ]] && fsck=1
	echo -e "UUID=${UUID}\\t${MOUNTPOINT}\\t${FSFORMAT}\\t$options\\t0\\t$fsck" | tee -a $fstab >&2
}

setup_disks() {
	#shellcheck disable=SC2207
	disks=($(stormeta 'disks[].device'))
	msg '########## setting raids #########'
	raids=('hack')
	#shellcheck disable=SC2207
	fsdevs=($(stormeta 'filesystems[].mount.device'))
	if stormeta 'raid[0].devices[]' &>/dev/null; then
		#shellcheck disable=SC2207
		raids=($(stormeta 'raid[].name'))
		msg "RAID devs: ${raids[*]}"
	else
		msg "No raid defined in config json - skipping raid config..."
	fi

	msg "Disks: ${disks[*]}"

	## Create partitions on each disk
	diskcnt=0
	for disk in "${disks[@]}"; do
		msg "Writing disk config for $disk"
		#TODO _maybe_ consider wipe/zap if wipeTable is defined for disk
		## Possibly key off just disks[0] to ensure symetrical layout for raid, but only if the bd is a raid member
		if [[ $preserve_data == true ]] || [[ $deprovision_fast == true ]]; then
			msg "Wiping disk partition table due to preserve_data and/or deprovision_fast"
			fast_wipe "$disk"
		fi

		#shellcheck disable=SC2207
		parts=($(stormeta "disks[$diskcnt].partitions[].number"))
		for partcnt in $(seq 0 $((${#parts[@]} - 1))); do
			partlabel=$(stormeta "disks[$diskcnt].partitions[$partcnt].label")
			partnum=$(stormeta "disks[$diskcnt].partitions[$partcnt].number")
			partsize=$(stormeta "disks[$diskcnt].partitions[$partcnt].size")
			parttype=8300
			msg "Working on $disk part $partnum aka label $partlabel with size $partsize"

			if [[ $partlabel =~ "BIOS" ]]; then
				bootdevs+=("$disk")
				case ${UEFI:-} in
				true)
					humantype="EFI system"
					parttype=ef00
					;;
				*)
					humantype="BIOS boot partition"
					parttype=ef02
					;;
				esac
				msg "label contains BIOS, setting partition type as $humantype"
			fi

			if [[ $partsize =~ "pct" ]]; then
				msg "Part size contains pct. Lets do some math"
				# all maths is done on sector basis
				disksize=$(blockdev --getsz "$disk")
				partpct=${partsize//pct/}
				npartsize=$((disksize * partpct / 100))

				msg "Creating a partition $partpct% the size of disk $disk ($disksize) resulting in partition size $npartsize sectors"
				partsize=+$npartsize
			elif ((partsize == 0)); then
				msg "Partsize is zero. This means use the rest of the disk."
			else
				partsize="+$partsize"
			fi
			sgdisk -n "$partnum:0:$partsize" -c "$partnum:$partlabel" -t "$partnum:$parttype" "$disk" >&2
		done
		diskcnt=$((diskcnt + 1))
	done

	## Create RAID array(s)
	raidcnt=0
	for raid in "${raids[@]}"; do
		# older versions of bash treat empty arrays as unset :arghfist:
		[[ $raid == 'hack' ]] && break

		msg "Writing RAID config for array $raid"
		# shellcheck disable=SC2207
		raiddevs=($(stormeta "raid[$raidcnt].devices[]"))
		raiddevscnt=${#raiddevs[@]}
		raidlevel=$(stormeta "raid[$raidcnt].level")
		msg "RAID dev list: ${raiddevs[*]}"
		msg "level: $raidlevel raid-devices: $raiddevscnt"
		mdadm --create "$raid" --force --run --level="$raidlevel" --raid-devices="$raiddevscnt" "${raiddevs[@]}" >&2

		raidcnt=$((raidcnt + 1))
	done

	## Create filesystems and update fstab
	fscnt=0
	for fsdev in "${fsdevs[@]}"; do
		msg "Writing filesystem for $fsdev"
		format=$(stormeta "filesystems[$fscnt].mount.format")
		mpoint=$(stormeta "filesystems[$fscnt].mount.point")
		# shellcheck disable=SC2207
		fsopts=($(stormeta "filesystems[$fscnt].mount.create.options[]"))
		fscnt=$((fscnt + 1))

		msg "$fsdev: format=$format fsopts=\"${fsopts[*]}\" mpoint=\"$mpoint\""
		if [[ $format == 'bios' ]]; then
			continue
		elif [[ $format == 'swap' ]]; then
			mkswap "$fsdev" >&2
		else
			"mkfs.$format" -F "${fsopts[@]}" "$fsdev" >&2
		fi

		thisdevuuid=$(blkid -s UUID -o value "$fsdev")
		if [[ ${mpoint} == / ]]; then
			rootuuid=$thisdevuuid

		fi

		add_to_fstab "${thisdevuuid}" "${mpoint}" "${format}"

	done

	# if this succeeds then we are guaranteed that bootdevs and rootuuid are set
	cat <<-EOF | python3
		import json
		d = {
		        'fstab': open("$fstab").read(),
		        'rootuuid': '$rootuuid',
		        'bootdevs': '${bootdevs[*]}'.split(),
		}
		print(json.dumps(d))
	EOF
}

setup_remount() {
	jq -r .fstab "$1" | tee $fstab
}

case $# in
4) setup_disks ;;
6)
	[[ $5 != 'mount' ]] && echo "cpr.sh: expected 5th arg to be 'mount'" && exit 1
	setup_remount "$6"
	;;
*) echo "cpr.sh called incorrectly" && exit 1 ;;
esac

target=/mnt/target
fstab=/tmp/fstab.tmpl

msg "${GREEN}#### Mounting filesystems in fstab to $target"
sed -i 's|\(\S\)\t/|\1\t'"$target"'/|' $fstab
while read -r mnt; do
	#msg "$mnt"
	mkdir -p "$mnt"
	mount --fstab "$fstab" "$mnt"
done < <(awk '/\/mnt\// {print $2}' $fstab | sort)
