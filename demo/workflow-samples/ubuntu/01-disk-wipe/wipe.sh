#!/bin/bash

source functions.sh && init
set -o nounset

# defaults
# shellcheck disable=SC2207
disks=($(lsblk -dno name -e1,7,11 | sed 's|^|/dev/|' | sort))

stimer=$(date +%s)

# Look for active MD arrays
# shellcheck disable=SC2207
mdarrays=($(awk '/md/ {print $4}' /proc/partitions))
if ((${#mdarrays[*]} != 0)); then
        for mdarray in "${mdarrays[@]}"; do
                echo "MD array: $mdarray"
                mdadm --stop "/dev/$mdarray"
                # sometimes --remove fails, according to manpages seems we
                # don't need it / are doing it wrong
                mdadm --remove "/dev/$mdarray" || :
        done
fi

# Wipe the filesystem and clear block on each block device
for bd in "${disks[@]}"; do
        sgdisk -Z "$bd" &
done
for bd in "${disks[@]}"; do
        # -n is so that wait will return on any job finished, returning it's exit status.
        # Without the -n, wait will only report exit status of last exited process
        wait -n
done

if [[ -d /sys/firmware/efi ]]; then
        for bootnum in $(efibootmgr | sed -n '/^Boot[0-9A-F]/ s|Boot\([0-9A-F]\{4\}\).*|\1|p'); do
                efibootmgr -Bb "$bootnum"
        done
fi

echo "Disk wipe finished."

## End installation
etimer=$(date +%s)
echo -e "${BYELLOW}Clean time: $((etimer - stimer))${NC}"
