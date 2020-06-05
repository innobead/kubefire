#!/usr/bin/env bash

zypper ref

zypper in -f -y ca-certificates-suse

zypper in -f -y systemd
ln -s /usr/lib/systemd/systemd /sbin/init

zypper -n install -f -y \
        iproute2 \
        iputils \
        openssh \
        net-tools \
        systemd-sysvinit \
        udev \
        sudo \
        wget \
        e2fsprogs \
        device-mapper

zypper install -t pattern -f -y apparmor
zypper install -f -y apparmor-utils

echo "root:root" | chpasswd

sed -i -E "s/#PasswordAuthentication no/PasswordAuthentication no/g" /etc/ssh/sshd_config
systemctl enable sshd

zypper clean --all