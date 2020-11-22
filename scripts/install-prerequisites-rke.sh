#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail
set -o xtrace

TMP_DIR=/tmp/kubefire
RKE_VERSION=${RKE_VERSION:-}

ARCH=$(uname -m)
case $ARCH in
"x86_64")
  ARCH="amd64"
  ;;
*)
  echo "Unsupported architecture ${ARCH}" >/dev/stderr
  exit 1
  ;;
esac

if [ -z "$RKE_VERSION" ]; then
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

function install_rke() {
  curl -sfSL "https://github.com/rancher/rke/releases/download/${RKE_VERSION}/rke_linux-${ARCH}" -o rke
  chmod +x rke && sudo mv rke /usr/local/bin/
}

install_rke
