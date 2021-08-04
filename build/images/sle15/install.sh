#!/usr/bin/env bash

set -o errexit

cleanup() {
  zypper rr product
  zypper rr product-update
  zypper rr basesystem
  zypper rr basesystem-update
}
trap cleanup EXIT

zypper ar http://download.suse.de/ibs/SUSE/Products/SLE-Product-SLES/15-SP3/x86_64/product product
zypper ar http://download.suse.de/ibs/SUSE/Updates/SLE-Product-SLES/15-SP3/x86_64/update/ product-update
zypper ar http://download.suse.de/ibs/SUSE/Products/SLE-Module-Basesystem/15-SP3/x86_64/product/ basesystem
zypper ar http://download.suse.de/ibs/SUSE/Updates/SLE-Module-Basesystem/15-SP3/x86_64/update/ basesystem-update

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
        SUSEConnect \
        tar \
        curl \
        ethtool \
        socat \
        ebtables \
        iptables \
        conntrack-tools

zypper install -t pattern -f -y apparmor
zypper install -f -y apparmor-utils

zypper rm -y container-suseconnect

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
