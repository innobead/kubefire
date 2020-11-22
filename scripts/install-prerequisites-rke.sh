#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail
set -o xtrace

function install_rke() {
  curl -sfSL "https://github.com/rancher/rke/releases/download/${RKE_VERSION}/rke_linux-${ARCH}" -o rke
  chmod +x rke && sudo mv rke /usr/local/bin/
}

function install_docker() {
  if [[ $(command -v apt) ]]; then
    sudo apt-get update
    sudo apt install -y docker.io
    sudo systemctl start docker
    sudo systemctl enable docker

  elif [[ $(command -v zypper) ]]; then
    sudo zypper install -y docker
    sudo systemctl start docker
    sudo systemctl enable docker

  elif [[ $(command -v dnf) ]]; then
    sudo dnf config-manager --add-repo=https://download.docker.com/linux/centos/docker-ce.repo
    sudo dnf install docker-ce --nobest -y
    sudo systemctl start docker
    sudo systemctl enable docker

  else
    echo "Unable to install docker, unsupported package manager" >/dev/stderr
    exit 1
  fi
}

# for nodes
if [[ $# -eq 1 ]]; then
  install_docker
  exit 0
fi

# for host
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

install_rke