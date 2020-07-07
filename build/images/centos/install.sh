#!/usr/bin/env bash

yum -y install \
  iproute \
  iputils \
  openssh-server \
  openssh-clients \
  net-tools \
  procps-ng \
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

yum clean all

echo "root:root" | chpasswd

sed -i -E "s/#PasswordAuthentication no/PasswordAuthentication no/g" /etc/ssh/sshd_config
systemctl enable sshd

cat <<'EOF' >>/etc/profile
if [ ! -S ~/.ssh/ssh_auth_sock ]; then
  eval `ssh-agent`
  ln -sf "$SSH_AUTH_SOCK" ~/.ssh/ssh_auth_sock
fi
export SSH_AUTH_SOCK=~/.ssh/ssh_auth_sock
ssh-add -l > /dev/null || ssh-add
EOF
