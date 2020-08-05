#!/usr/bin/env bash
# ref: https://ignite.readthedocs.io/en/stable/installation/

set -o errexit
set -o nounset
set -o pipefail
set -o xtrace

TMP_DIR=/tmp/kubefire
GOARCH=$(go env GOARCH 2>/dev/null || echo "amd64")

# FIXME: below versions should come from kubefire, remove the versions in near future
CONTAINERD_VERSION=${CONTAINERD_VERSION:-"v1.3.4"}
IGNITE_VERION=${IGNITE_VERION:-"v0.7.1"}
CNI_VERSION=${CNI_VERSION:-"v0.8.6"}
RUNC_VERSION=${RUNC_VERSION:-"v1.0.0-rc91"}

mkdir -p $TMP_DIR
pushd $TMP_DIR

function cleanup() {
  rm -rf $TMP_DIR || true
}

trap cleanup EXIT ERR INT TERM

function _check_version() {
  set +o pipefail

  local exec_name=$1
  local exec_version_cmd=$2
  local version=$3

  command -v "${exec_name}" && [[ "$(eval "$exec_name $exec_version_cmd 2>&1")" =~ $version ]]
  return $?
}

function check_virtualization() {
  lscpu | grep Virtualization
  lsmod | grep kvm
}

function install_containerd() {
  if _check_version /usr/local/bin/containerd --version $CONTAINERD_VERSION; then
    echo "containerd (${CONTAINERD_VERSION}) installed already!"
    return
  fi

  local version="${CONTAINERD_VERSION:1}"
  local dir=containerd-$version

  curl -sSLO "https://github.com/containerd/containerd/releases/download/${CONTAINERD_VERSION}/containerd-${version}.linux-${GOARCH}.tar.gz"
  mkdir -p $dir
  tar -zxvf $dir*.tar.gz -C $dir
  chmod +x $dir/bin/*
  sudo mv $dir/bin/* /usr/local/bin/

  curl -sSLO "https://raw.githubusercontent.com/containerd/containerd/${CONTAINERD_VERSION}/containerd.service"
  sudo groupadd containerd || true
  sudo mv containerd.service /etc/systemd/system/containerd.service

#  [Service]
#  ExecStartPre=-/sbin/modprobe overlay
#  ExecStart=/usr/local/bin/containerd
#  ExecStartPost=/usr/bin/chgrp containerd /run/containerd/containerd.sock

  sudo mkdir -p /etc/containerd
  containerd config default | sudo tee /etc/containerd/config.toml >/dev/null
  sudo systemctl enable --now containerd
}

function install_runc() {
  if _check_version /usr/local/bin/runc -version $RUNC_VERSION; then
    echo "runc (${RUNC_VERSION}) installed already!"
    return
  fi

  curl -sSL "https://github.com/opencontainers/runc/releases/download/${RUNC_VERSION}/runc.amd64" -o runc
  chmod +x runc
  sudo mv runc /usr/local/bin/
}

function install_cni() {
  if _check_version /opt/cni/bin/bridge --version $CNI_VERSION; then
    echo "CNI plugins (${CNI_VERSION}) installed already!"
    return
  fi

  mkdir -p /opt/cni/bin
  curl -L "https://github.com/containernetworking/plugins/releases/download/${CNI_VERSION}/cni-plugins-linux-amd64-${CNI_VERSION}.tgz" | tar -C /opt/cni/bin -xz
}

function install_ignite() {
  if _check_version /usr/local/bin/ignite version $IGNITE_VERION; then
    echo "ignite (${IGNITE_VERION}) installed already!"
    return
  fi

  for binary in ignite ignited; do
    echo "Installing $binary..."
    curl -sfLo $binary "https://github.com/weaveworks/ignite/releases/download/${IGNITE_VERION}/${binary}-${GOARCH}"
    chmod +x $binary
    sudo mv $binary /usr/local/bin
  done
}

function check_ignite() {
  ignite version
}

check_virtualization

install_runc
install_containerd
install_cni
install_ignite
check_ignite

popd
