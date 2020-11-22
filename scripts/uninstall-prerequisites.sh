#!/usr/bin/env bash

set -o errexit
set -o nounset
#set -o pipefail
#set -o xtrace

function _is_arm_arch() {
    uname -m | grep "aarch64"
    return $?
}

function ask() {
  read -p "Uninstall $1, are you sure [Y/y]? " -n 1 -r
  echo
  if [[ $REPLY =~ ^[Yy]$ ]]; then
    # shellcheck disable=SC2086
    uninstall_$1
  else
    echo "Ignored to uninstall $1"
  fi
}

function uninstall_containerd() {
  if _is_arm_arch; then
    echo "!!! Please uninstall containerd aarch64 via system package manager !!!"
    return
  fi

  sudo rm -f /usr/local/bin/containerd* /usr/local/bin/ctr
}

function uninstall_runc() {
  if _is_arm_arch; then
    echo "!!! Please uninstall containerd aarch64 via system package manager !!!"
    return
  fi

  sudo rm -f /usr/local/bin/runc
}

function uninstall_cni() {
  sudo rm -rf /opt/cni/bin
  sudo rm -rf /var/lib/cni/networks/kubefire-cni-bridge
}

function uninstall_ignite() {
  # shellcheck disable=SC2046
  sudo ignite rm -f $(sudo ignite ps -aq) &>/dev/null || true
  sudo ignite image ls -q | xargs sudo ignite image rm &>/dev/null || true
  sudo ignite kernel ls -q | xargs sudo ignite kernel rm &>/dev/null || true

  # clean up firecracker vms, kernels, rootfs
  sudo rm -rf /var/lib/firecracker

  # clean up ignite executables
  sudo rm -f /usr/local/bin/ignite{,d}
}

ask containerd
ask runc
ask cni
ask ignite
