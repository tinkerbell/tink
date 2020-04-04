#!/bin/bash

source functions.sh && init
set -o nounset

# defaults
# shellcheck disable=SC2207

#metadata=/metadata
#curl -sSL --connect-timeout 60 https://metadata.packet.net/metadata > $metadata
#check_required_arg "$metadata" 'metadata file' '-M'
#declare class && set_from_metadata class 'class' <"$metadata"
#declare facility && set_from_metadata facility 'facility' <"$metadata"
#declare os && set_from_metadata os 'operating_system.slug' <"$metadata"
#declare tag && set_from_metadata tag 'operating_system.image_tag' <"$metadata" || tag=""

ephemeral=/workflow/data.json
class=$(jq -r .class "$ephemeral")
facility=$(jq -r .facility "$ephemeral")
os=$(jq -r .os "$ephemeral")
tag=$(jq -r .tag "$ephemeral")

OS=$os${tag:+:$tag}
arch=$(uname -m)
custom_image=false
target="/mnt/target"
cprout=/statedir/cpr.json

# custom
mkdir -p $target
mkdir -p $target/boot
mount -t efivarfs efivarfs /sys/firmware/efi/efivars
mount -t ext4 /dev/sda3 $target
mkdir -p $target/boot/efi

if ! [[ -f /statedir/disks-partioned-image-extracted ]]; then
    assetdir=/tmp/assets
    mkdir $assetdir
    OS=${OS%%:*}

    # Grub config
    BASEURL="http://$MIRROR_HOST/misc/osie/current"
    
    # Ensure critical OS dirs
    mkdir -p $target/{dev,proc,sys}
    mkdir -p $target/etc/mdadm

    if [[ $class != "t1.small.x86" ]]; then
        echo -e "${GREEN}#### Updating MD RAID config file ${NC}"
        mdadm --examine --scan >>$target/etc/mdadm/mdadm.conf
    fi

    # ensure unique dbus/systemd machine-id
    echo -e "${GREEN}#### Setting machine-id${NC}"
    rm -f $target/etc/machine-id $target/var/lib/dbus/machine-id

    systemd-machine-id-setup --root=$target
    cat $target/etc/machine-id
    [[ -d $target/var/lib/dbus ]] && ln -nsf /etc/machine-id $target/var/lib/dbus/machine-id

    # Install kernel and initrd
    # Kernel to throw on the target
    kernel="$assetdir/kernel.tar.gz"
    # Initrd to throw on the target
    initrd="$assetdir/initrd.tar.gz"
    # Modules to throw on the target
    modules="$assetdir/modules.tar.gz"
    echo -e "${GREEN}#### Fetching and copying kernel, modules, and initrd to target $target ${NC}"
    wget "$BASEURL/$os/kernel.tar.gz" -P $assetdir
    wget "$BASEURL/$os/initrd.tar.gz" -P $assetdir
    wget "$BASEURL/$os/modules.tar.gz" -P $assetdir
    tar --warning=no-timestamp -zxf "$kernel" -C $target/boot
    kversion=$(vmlinuz_version $target/boot/vmlinuz)
    if [[ -z $kversion ]]; then
        echo 'unable to extract kernel version' >&2
        exit 1
    fi

    kernelname="vmlinuz-$kversion"
    if [[ ${os} =~ ^centos ]] || [[ ${os} =~ ^rhel ]]; then
            initrdname=initramfs-$kversion.img
            modulesdest=usr
    else
            initrdname=initrd.img-$kversion
            modulesdest=
    fi

    mv $target/boot/vmlinuz "$target/boot/$kernelname" && ln -nsf "$kernelname" $target/boot/vmlinuz
    tar --warning=no-timestamp -zxf "$initrd" && mv initrd "$target/boot/$initrdname" && ln -nsf "$initrdname" $target/boot/initrd
    tar --warning=no-timestamp -zxf "$modules" -C "$target/$modulesdest"
    cp "$target/boot/$kernelname" /statedir/kernel
    cp "$target/boot/$initrdname" /statedir/initrd

    # Install grub
    #grub="$BASEURL/grub/${OS//_(arm|image)//}/$class/grub.template"
    grub="$BASEURL/grub/$os/$class/grub.template"
    echo -e "${GREEN}#### Installing GRUB2${NC}"

    wget "$grub" -O /tmp/grub.template
    wget "${grub}.default" -O /tmp/grub.default

    mkfs.vfat -c -F 32 /dev/sda1
    mount -t vfat /dev/sda1 $target/boot/efi
    ./grub-installer.sh -v -p "$class" -t "$target" -C "$cprout" -D /tmp/grub.default -T /tmp/grub.template

    rootuuid=$(jq -r .rootuuid $cprout)
    [[ -n $rootuuid ]]
    cmdline=$(sed -nr 's|GRUB_CMDLINE_LINUX='\''(.*)'\''|\1|p' /tmp/grub.default)

    echo -e "${GREEN}#### Clearing init overrides to enable TTY${NC}"
    rm -rf $target/etc/init/*.override

    if [[ $custom_image == false ]]; then
            echo -e "${GREEN}#### Setting up package repos${NC}"
            ./repos.sh -a "$arch" -t $target -f "$facility" -M "$ephemeral"
    fi
fi

