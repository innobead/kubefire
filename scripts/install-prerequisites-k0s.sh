#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail
set -o xtrace

TMP_DIR=/tmp/kubefire

K0S_VERSION=${K0S_VERSION:-}
K0S_CONFIG=${K0S_CONFIG:-}
ARCH=${ARCH:-}

if [ -z "$K0S_VERSION" ]; then
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

function install_k0s() {
  local url="https://github.com/k0sproject/k0s/releases/download/${K0S_VERSION}/k0s-${K0S_VERSION}"

  ARCH=$(uname -m)
  case $ARCH in
  x86_64)
    url="$url"-amd64
    ;;
  aarch64)
    url="$url"-arm64
    ;;
  *)
    echo "not supported arch ${ARCH}" >/dev/stderr
    exit 1
    ;;
  esac

  curl -sfSL "$url" -o k0s
  chmod +x k0s && sudo mv k0s /usr/local/bin/
}

function create_config() {
  if [ -n "$K0S_CONFIG" ]; then
    mkdir -p /etc/k0s || true
    echo "$K0S_CONFIG" >/etc/k0s/config.yaml
  fi
}

$1
