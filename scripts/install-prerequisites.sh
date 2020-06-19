#!/usr/bin/env bash
# ref: https://ignite.readthedocs.io/en/stable/installation/

set -o errexit
set -o nounset
set -o pipefail
set -o xtrace

TMP_DIR=/tmp/kubefire
GOARCH=$(go env GOARCH 2>/dev/null || echo "amd64")
CONTAINERD_VERSION="v1.3.4"
IGNITE_VERION="v0.7.0"
CNI_VERSION="v0.8.2"
RUNC_VERSION="v1.0.0-rc90"

mkdir -p $TMP_DIR
pushd $TMP_DIR

function cleanup() {
  rm -rf $TMP_DIR || true
}

trap cleanup EXIT ERR INT TERM

function check_virtualization() {
  lscpu | grep Virtualization
  lsmod | grep kvm
}

function check_container_runtime() {
  set +o pipefail

  ! pgrep containerd || pgrep dockerd
  return $?
}

function install_containerd() {
  local version="${CONTAINERD_VERSION:1}"
  local dir=containerd-$version

  curl -sSLO "https://github.com/containerd/containerd/releases/download/$CONTAINERD_VERSION/containerd-$version.linux-$GOARCH.tar.gz"
  mkdir -p $dir
  tar -zxvf $dir*.tar.gz -C $dir
  chmod +x $dir/bin/*
  sudo mv $dir/bin/* /usr/local/bin/

  curl -sSL "https://raw.githubusercontent.com/containerd/containerd/$CONTAINERD_VERSION/containerd.service" >/etc/systemd/system/containerd.service
  mkdir -p /etc/containerd
  containerd config default >/etc/containerd/config.toml
  systemctl enable --now containerd
}

function install_runc() {
  curl -sSL "https://github.com/opencontainers/runc/releases/download/$RUNC_VERSION/runc.amd64" -o runc
  chmod +x runc
  sudo mv runc /usr/local/bin/
}

function install_cni() {
  mkdir -p /opt/cni/bin
  curl -L "https://github.com/containernetworking/plugins/releases/download/${CNI_VERSION}/cni-plugins-linux-amd64-${CNI_VERSION}.tgz" | tar -C /opt/cni/bin -xz
}

function install_ignite() {
  for binary in ignite ignited; do
    echo "Installing $binary..."
    curl -sfLo $binary "https://github.com/weaveworks/ignite/releases/download/$IGNITE_VERION/$binary-$GOARCH"
    chmod +x $binary
    sudo mv $binary /usr/local/bin
  done
}

function check_ignite() {
  ignite version
}

check_virtualization

if ! check_container_runtime; then
  install_runc
  install_containerd
fi

install_cni
install_ignite
check_ignite

popd
