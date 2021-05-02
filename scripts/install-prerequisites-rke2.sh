#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail
set -o xtrace

TMP_DIR=/tmp/kubefire

RKE2_VERSION=${RKE2_VERSION:-}
RANCHERD_VERSION=${RANCHERD_VERSION:-}
RKE2_CONFIG=${RKE2_CONFIG:-}

if [ -z "$RKE2_VERSION" ]; then
  echo "incorrect versions provided!" >/dev/stderr
  exit 1
fi

rm -rf $TMP_DIR && mkdir -p $TMP_DIR
pushd $TMP_DIR

function cleanup() {
  rm -rf $TMP_DIR || true
  popd
}

trap cleanup EXIT ERR INT TERM

function install_rke2() {
  # https://get.rke2.io
  local url="https://raw.githubusercontent.com/rancher/rke2/${RKE2_VERSION}/install.sh"
  curl -sfSL "$url" -o rke2-install.sh
  chmod +x rke2-install.sh && sudo mv rke2-install.sh /usr/local/bin/
}

function install_rancherd() {
  local url="https://raw.githubusercontent.com/rancher/rancher/${RANCHERD_VERSION}/cmd/rancherd/install.sh"
  curl -sfSL "$url" -o rancherd-install.sh
  chmod +x rancherd-install.sh && sudo mv rancherd-install.sh /usr/local/bin/
}

function create_config() {
  if [ -n "$RKE2_CONFIG" ]; then
    mkdir -p /etc/rancher/rke2 || true
    echo "$RKE2_CONFIG" >/etc/rancher/rke2/config.yaml
  fi
}

$1
