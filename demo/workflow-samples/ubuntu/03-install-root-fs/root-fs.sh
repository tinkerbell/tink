#!/bin/bash

source functions.sh && init
set -o nounset

ephemeral=/workflow/data.json
os=$(jq -r .os "$ephemeral")
tag=$(jq -r .tag "$ephemeral")

target="/mnt/target"
OS=$os${tag:+:$tag}

if ! [[ -f /statedir/disks-partioned-image-extracted ]]; then
    ## Fetch install assets via git
    assetdir=/tmp/assets
    mkdir $assetdir
    echo -e "${GREEN}#### Fetching image for ${OS//_(arm|image)} root fs  ${NC}"
    OS=${OS%%:*}

    # Image rootfs
    image="$assetdir/image.tar.gz"

    # TODO: should come as ENV
    BASEURL="http://$MIRROR_HOST/misc/osie/current"

    # custom
    wget "$BASEURL/$os/image.tar.gz" -P $assetdir

    mkdir -p $target
    mount -t ext4 /dev/sda3 $target
    echo -e "${GREEN}#### Retrieving image and installing to target $target ${NC}"
    tar --xattrs --acls --selinux --numeric-owner --same-owner --warning=no-timestamp -zxpf "$image" -C $target
    echo -e "${GREEN}#### Success installing root fs ${NC}"   
fi

