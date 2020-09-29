#!/usr/bin/env bash

set -o errexit

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
        which \
        e2fsprogs \
        device-mapper \
        tar \
        curl \
        ethtool \
        socat \
        ebtables \
        iptables \
        conntrack-tools

zypper install -t pattern -f -y apparmor
zypper install -f -y apparmor-utils

zypper clean --all

echo "root:root" | chpasswd

sed -i -E "s/#PasswordAuthentication no/PasswordAuthentication no/g" /etc/ssh/sshd_config
systemctl enable sshd

cat <<'EOF' >> /etc/profile
if [ ! -S ~/.ssh/ssh_auth_sock ]; then
  eval `ssh-agent`
  ln -sf "$SSH_AUTH_SOCK" ~/.ssh/ssh_auth_sock
fi
export SSH_AUTH_SOCK=~/.ssh/ssh_auth_sock
ssh-add -l > /dev/null || ssh-add
EOF
