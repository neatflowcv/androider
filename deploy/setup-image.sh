#!/bin/bash
set -euo pipefail

echo "root 권한 확인"
if [ $(id -u) -ne 0 ]; then
    echo "root 권한이 없습니다. sudo 명령어를 사용해주세요."
    exit 1
fi

echo "insmod nbd"
if ! lsmod | grep -q '^nbd'; then
    modprobe nbd max_part=8
fi

echo "create qcow2 image"
qemu-img create -f qcow2 archlinux.qcow2 10G

echo "bind qcow2 image"
qemu-nbd -c /dev/nbd0 archlinux.qcow2

echo "partition qcow2 image"
parted /dev/nbd0 --script mklabel gpt
parted /dev/nbd0 --script mkpart ESP fat32 1MiB 512MiB
parted /dev/nbd0 --script set 1 boot on
parted /dev/nbd0 --script mkpart primary ext4 512MiB 100%

echo "format partition"
mkfs.fat -F32 /dev/nbd0p1
mkfs.ext4 -L rootfs /dev/nbd0p2

echo "mount partition"
mount /dev/nbd0p2 /mnt
mkdir -p /mnt/boot
mount /dev/nbd0p1 /mnt/boot

pacstrap /mnt base sudo linux linux-firmware mkinitcpio networkmanager iptables-nft --needed

genfstab -U /mnt > /mnt/etc/fstab

tee /mnt/etc/mkinitcpio.d/linux.preset > /dev/null <<EOF
ALL_kver="/boot/vmlinuz-linux"
PRESETS=('default' 'fallback')
default_uki="/boot/EFI/Linux/arch-linux.efi"
fallback_uki="/boot/EFI/Linux/arch-linux-fallback.efi"
EOF

mkdir -p /mnt/etc/cmdline.d
tee /mnt/etc/cmdline.d/root.conf > /dev/null <<EOF
root=LABEL=rootfs rw
EOF

arch-chroot /mnt bootctl install

arch-chroot /mnt mkinitcpio -p linux

tee /mnt/boot/loader/entries/arch.conf > /dev/null <<EOF
title   Arch Linux
efi /EFI/Linux/arch-linux.efi
options root=LABEL="rootfs" rw quiet splash console=tty0
EOF

echo "setting user"
echo root:root | arch-chroot /mnt chpasswd

echo "enable network manager"
arch-chroot /mnt systemctl enable NetworkManager

echo "unmount partition"
umount --recursive /mnt

echo "unbind qcow2 image"
qemu-nbd -d /dev/nbd0