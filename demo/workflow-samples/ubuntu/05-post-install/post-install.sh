#!/bin/bash

# TODO: ephemeral data or environment 
target="/mnt/target"
BASEURL='http://192.168.1.2/misc/osie/current'

# Adjust failsafe delays for first boot delay
if [[ -f $target/etc/init/failsafe.conf ]]; then
    sed -i 's/sleep 59/sleep 10/g' $target/etc/init/failsafe.conf
    sed -i 's/Waiting up to 60/Waiting up to 10/g' $target/etc/init/failsafe.conf
fi

wget $BASEURL/ubuntu_18_04/packet-block-storage-attach -P /
wget $BASEURL/ubuntu_18_04/packet-block-storage-detach -P /

echo -e "${GREEN}#### Run misc post-install tasks${NC}"
install -m755 -o root -g root /packet-block-storage-* $target/usr/bin
if [ -f $target/usr/sbin/policy-rc.d ]; then
    echo "Removing policy-rc.d from target OS."
    rm -f $target/usr/sbin/policy-rc.d
fi

