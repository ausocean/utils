#!/bin/bash
#
# Given a raspi image this script will adjust the partition layout
# so the root filesystem is mounted read only, and a /var filesystem is
# created to hold volatile data.
#
# The root filesystem will be set to just under 5GB, with /var taking the
# rest of the SD card.
#
# Joel Stanley <joel@jms.id.au>, April 2019

set -e

IMAGE="$1"
DEVICE="$2"
ROOTSIZE="5GB"
ROOTPART="$DEVICE"2
VARPART="$DEVICE"3
MOUNTPOINT=/tmp/raspi

if [ "$#" -ne 2 ]; then
    echo "usage: format-sd-card.sh [raspi image] [block device]"
    return 1
fi

echo "Image $IMAGE"
echo ""
echo "Using $DEVICE as device. This operation is destructive. ctrl-c to abort"
echo ""

# Time to bail
sleep 5

echo "Checking for tools..."
command -v dd
command -v parted
command -v mkfs.ext4
command -v e2fsck
command -v resize2fs
command -v blkid
echo ""

echo "Copying image to SD card. This will take a while..."
echo "(run 'sudo kill -USR1 \`pgrep dd\` for progress)"
dd if="$IMAGE" of="$DEVICE"

echo "Resizing root partition..."
parted "$DEVICE" resizepart 2 $ROOTSIZE

echo "Resize root filesystem..."
e2fsck -f "$ROOTPART"
resize2fs "$ROOTPART"

echo "Creating /var partition..."
parted "$DEVICE" mkpart primary ext4 $ROOTSIZE 100%

echo "Create /var filesystem..."
mkfs.ext4 "$VARPART"

echo "Mounting root filesystem..."
mkdir -p "$MOUNTPOINT"
mount "$ROOTPART" "$MOUNTPOINT"

if [ ! -f "$MOUNTPOINT"/etc/rpi-issue ]; then
	echo "Doesn't look like a raspi filesystem, aborting"
	return 1
fi

echo "Adding 'ro' to /boot and / partitions..."
sed -i '2,3s/defaults/defaults,ro/' "$MOUNTPOINT"/etc/fstab

echo "Adding /var to fstab..."
VARID=$(blkid -o value -s PARTUUID "$VARPART")
printf "PARTUUID=$VARID  /var\text4\tdefaults,noatime\t0\t2\n" >> "$MOUNTPOINT"/etc/fstab

umount "$MOUNTPOINT"
rmdir "$MOUNTPOINT"
