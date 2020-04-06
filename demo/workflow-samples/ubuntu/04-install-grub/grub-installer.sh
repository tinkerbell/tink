#!/bin/bash

source functions.sh && init

USAGE="Usage: $0 -t /mnt/target -C /path/to/cprout.json
Required Arguments:
	-p plan      Server plan (ex: t1.small.x86)
	-t target    Target mount point to write configs to
	-C path      Path to file containing cpr.sh output json
	-D path      Path to grub.default template
	-T path      Path to grub.cfg template

Options:
	-h           This help message
	-v           Turn on verbose messages for debugging

Description: This script will configure grub for the target distro
"
while getopts "p:t:C:D:T:hv" OPTION; do
	case $OPTION in
	p) plan=$OPTARG ;;
	t) target="$OPTARG" ;;
	C) cprout=$OPTARG ;;
	D) default_path=$OPTARG ;;
	T) template_path=$OPTARG ;;
	h) echo "$USAGE" && exit 0 ;;
	v) set -x ;;
	*) echo "$USAGE" && exit 1 ;;
	esac
done

assert_all_args_consumed "$OPTIND" "$@"

# Make sure target provided is mounted
if grep -qs "$target" /proc/mounts; then
	echo "Target is mounted... good."
else
	echo "Error: Target $target is not mounted"
	exit 1
fi

rm -rf "$target/boot/grub"
[[ -d $target/boot/grub2 ]] || mkdir -p "$target/boot/grub2"
ln -nsfT grub2 "$target/boot/grub"

rootuuid=$(jq -r .rootuuid "$cprout")
[[ -n $rootuuid ]]
sed "s/PACKET_ROOT_UUID/$rootuuid/g" "$template_path" >"$target/boot/grub2/grub.cfg"

cmdline=$(sed -nr 's|GRUB_CMDLINE_LINUX='\''(.*)'\''|\1|p' "$default_path")
echo -e "${BYELLOW}Detected cmdline: ${cmdline}${NC}"
(
	sed -i 's|^|export |' "$default_path"
	# shellcheck disable=SC1090
	# shellcheck disable=SC1091
	source "$default_path"
	GRUB_DISTRIBUTOR=$(detect_os "$target" | awk '{print $1}') envsubst <grub.default >"$target/etc/default/grub"
)

is_uefi && uefi=true || uefi=false
arch=$(uname -m)
os_ver=$(detect_os "$target")

# shellcheck disable=SC2086
set -- $os_ver
DOS=$1
DVER=$2
echo "#### Detected OS on mounted target $target"
echo "OS: $DOS  ARCH: $arch VER: $DVER"

chroot_install=false

if [[ $DOS == "RedHatEnterpriseServer" ]] && [[ $arch == "aarch64" ]] || [[ $plan == "c3.medium.x86" ]]; then
	chroot_install=true
fi

install_grub_chroot() {
	echo "Attempting to install Grub on $disk"
	mount --bind /dev "$target/dev"
	mount --bind /tmp "$target/tmp"
	mount --bind /proc "$target/proc"
	mount --bind /sys "$target/sys"
	chroot "$target" /bin/bash -xe <<EOF
is_uefi=false
[[ -d /sys/firmware/efi ]] && {
	is_uefi=true
	mountpoint -q /sys/firmware/efi/efivars || {
		mount -t efivarfs efivarfs /sys/firmware/efi/efivars
	}
}

if which grub2-install; then
	grub2-install --recheck "$disk"
elif which grub-install; then
	grub-install --recheck "$disk"
else
	echo 'grub-install or grub2-install are not installed on target os'
	exit 1
fi
\$is_uefi && {
	[ -f /etc/os-release ] && {
		(
			source /etc/os-release
			efibootmgr | tee /dev/stderr | grep -iq "\$ID"
		)
	}
	umount /sys/firmware/efi/efivars
}
EOF
	umount "$target/dev" "$target/tmp" "$target/proc" "$target/sys"

	if $uefi && [[ $plan == "c3.medium.x86" ]] && [[ $DOS == "CentOS" ]]; then
		add_post_install_service "$target"
		efi_uuid="$(efi_device "$target/boot/efi" -o partuuid)"
		efi_id="$(find_uuid_boot_id "$efi_uuid")"
		echo "Forcing next boot to be target os"
		efibootmgr -n "$efi_id"
	fi
}

install_grub_osie() {
	echo "Running grub-install on $disk"
	if ! $uefi; then
		grub-install --recheck --root-directory="$target" "$disk"
	else
		[[ $arch == aarch64 ]] && mount -o remount,ro /sys/firmware/efi/efivars

		grub-install --recheck --bootloader-id=GRUB --root-directory="$target" --efi-directory="$target/boot/efi"
		grubefi=$(find "$target/boot/efi" -name 'grub*.efi' -print -quit)
		if [[ -z $grubefi ]]; then
			echo "error: couldn't find a suitable grub EFI file"
			exit 1
		fi

		if [[ $arch == aarch64 ]]; then
			echo "Renaming $grubefi to default BOOT binary"
			install -Dm755 "$grubefi" "$target/boot/efi/EFI/BOOT/BOOTAA64.EFI"
			install -Dm755 "$grubefi" "$target/boot/efi/EFI/GRUB2/GRUBAA64.EFI"
		else
			# grub-install doesn't error if efibootmgr can't actually set the boot entries/order
			efibootmgr | tee /dev/stderr | grep -iq grub
		fi
	fi
}

# shellcheck disable=SC2207
bootdevs=$(jq -r '.bootdevs[]' "$cprout")
[[ -n $bootdevs ]]
for disk in $bootdevs; do
	if $chroot_install; then
		install_grub_chroot
	else
		install_grub_osie
	fi
done
