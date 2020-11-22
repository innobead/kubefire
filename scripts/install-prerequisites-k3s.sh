#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail
set -o xtrace

TMP_DIR=/tmp/kubefire

K3S_VERSION=${K3S_VERSION:-}

if [ -z "$K3S_VERSION" ]; then
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

function install_k3s() {
  # https://get.k3s.io
  local url="https://raw.githubusercontent.com/rancher/k3s/${K3S_VERSION}+k3s1/install.sh" # for backward compatible
  if [[ "$K3S_VERSION" =~ .*+k3s.* ]]; then
    local url="https://raw.githubusercontent.com/rancher/k3s/${K3S_VERSION}/install.sh"
  fi

  curl -sfSL "$url" -o k3s-install.sh
  chmod +x k3s-install.sh && sudo mv k3s-install.sh /usr/local/bin/
}

install_k3s
