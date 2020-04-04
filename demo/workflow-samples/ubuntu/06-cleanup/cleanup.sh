#!/bin/bash

cprout=/statedir/cpr.json
rootuuid=$(jq -r .rootuuid $cprout)
[[ -n $rootuuid ]]
cmdline=$(sed -nr 's|GRUB_CMDLINE_LINUX='\''(.*)'\''|\1|p' /tmp/grub.default)

kexec -l ./kernel --initrd=./initrd --command-line="BOOT_IMAGE=/boot/vmlinuz root=UUID=$rootuuid ro $cmdline" || reboot
kexec -e || reboot

