#!/bin/bash

source functions.sh && init
set -o nounset

# defaults
# shellcheck disable=SC2207
disks=($(lsblk -dno name -e1,7,11 | sed 's|^|/dev/|' | sort))
userdata='/dev/null'

arch=$(uname -m)

metadata=/metadata
curl --connect-timeout 60 http://$MIRROR_HOST:50061/metadata > $metadata
check_required_arg "$metadata" 'metadata file' '-M'

declare class && set_from_metadata class 'plan_slug' <"$metadata"
declare facility && set_from_metadata facility 'facility_code' <"$metadata"
declare os && set_from_metadata os 'instance.operating_system_version.os_slug' <"$metadata"
declare preserve_data && set_from_metadata preserve_data 'preserve_data' false <"$metadata"

# declare pwhash && set_from_metadata pwhash 'password_hash' <"$metadata"
# declare state && set_from_metadata state 'state' <"$metadata"
declare pwhash="5f4dcc3b5aa765d61d8327deb882cf99"
declare state="provisioning"

declare tag && set_from_metadata tag 'instance.operating_system_version.image_tag' <"$metadata" || tag=""
#declare tinkerbell && set_from_metadata tinkerbell 'phone_home_url' <"$metadata"
declare deprovision_fast && set_from_metadata deprovision_fast 'deprovision_fast' false <"$metadata"

OS=$os${tag:+:$tag}
echo "Number of drives found: ${#disks[*]}"
if ((${#disks[*]} != 0)); then
        echo "Disk candidate check successful"
fi

ephemeral=/workflow/data.json
echo "{}" > $ephemeral
echo $(jq ". + {\"arch\": \"$arch\"}" <<< cat $ephemeral) > $ephemeral
echo $(jq ". + {\"class\": \"$class\"}" <<< cat $ephemeral) > $ephemeral
echo $(jq ". + {\"facility\": \"$facility\"}" <<< cat $ephemeral) > $ephemeral
echo $(jq ". + {\"os\": \"$os\"}" <<< cat $ephemeral) > $ephemeral
echo $(jq ". + {\"preserve_data\": \"$preserve_data\"}" <<< cat $ephemeral) > $ephemeral
echo $(jq ". + {\"pwhash\": \"$pwhash\"}" <<< cat $ephemeral) > $ephemeral
echo $(jq ". + {\"state\": \"$state\"}" <<< cat $ephemeral) > $ephemeral
echo $(jq ". + {\"tag\": \"$tag\"}" <<< cat $ephemeral) > $ephemeral
#echo $(jq ". + {\"tinkerbell\": \"$tinkerbell\"}" <<< cat $ephemeral) > $ephemeral
echo $(jq ". + {\"deprovision_fast\": \"$deprovision_fast\"}" <<< cat $ephemeral) > $ephemeral

jq . $ephemeral

custom_image=false
target="/mnt/target"
cprconfig=/tmp/config.cpr
cprout=/statedir/cpr.json

# custom
#mkdir -p /statedir && touch /statedir/cpr.json
touch /statedir/cpr.json

echo -e "${GREEN}#### Checking userdata for custom cpr_url...${NC}"
cpr_url=$(sed -nr 's|.*\bcpr_url=(\S+).*|\1|p' "$userdata")

if [[ -z ${cpr_url} ]]; then
        echo "Using default image since no cpr_url provided"
        jq -c '.instance.storage' "$metadata" >$cprconfig
else
        echo "NOTICE: Custom CPR url found!"
        echo "Overriding default CPR location with custom cpr_url"
        if ! curl "$cpr_url" | jq . >$cprconfig; then
                phone_home "${tinkerbell}" '{"instance_id":"'"$(jq -r .id "$metadata")"'"}'
                echo "$0: CPR URL unavailable: $cpr_url" >&2
                exit 1
        fi
fi

if ! [[ -f /statedir/disks-partioned-image-extracted ]]; then
        OS=${OS%%:*}
        jq . $cprconfig

        # make sure the disks are ok to use
        assert_block_or_loop_devs "${disks[@]}"
        assert_same_type_devs "${disks[@]}"

        is_uefi && uefi=true || uefi=false

        if [[ $deprovision_fast == false ]] && [[ $preserve_data == false ]]; then
                echo -e "${GREEN}Checking disks for existing partitions...${NC}"
                if fdisk -l "${disks[@]}" 2>/dev/null | grep Disklabel >/dev/null; then
                        echo -e "${RED}Critical: Found pre-exsting partitions on a disk. Aborting install...${NC}"
                        fdisk -l "${disks[@]}"
                        exit 1
                fi
        fi

        echo "Disk candidates are ready for partitioning."

        echo -e "${GREEN}#### Running CPR disk config${NC}"
        UEFI=$uefi ./cpr.sh $cprconfig "$target" "$preserve_data" "$deprovision_fast" | tee $cprout

        mount | grep $target

        # dump cpr provided fstab into $target
        mkdir -p /mnt/target/etc
        touch  /mnt/target/etc/fstab
        jq -r .fstab "$cprout" >$target/etc/fstab

        echo -e "${GREEN}#### CPR disk config complete ${NC}"
fi
