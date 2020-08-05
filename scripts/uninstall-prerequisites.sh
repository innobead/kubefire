#!/usr/bin/env bash

set -o errexit
set -o nounset
#set -o pipefail
#set -o xtrace

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
  sudo rm -f /usr/loca/bin/containerd* /usr/local/bin/ctr
}

function uninstall_runc() {
  sudo rm -f /usr/local/bin/runc
}

function uninstall_cni() {
  sudo rm -rf /opt/cni/bin
}

function uninstall_ignite() {
  # shellcheck disable=SC2046
  sudo ignite rm -f $(sudo ignite ps -aq)
  sudo rm -rf /var/lib/firecracker
  sudo rm -f /usr/local/bin/ignite{,d}
}

ask containerd
ask runc
ask cni
ask ignite
