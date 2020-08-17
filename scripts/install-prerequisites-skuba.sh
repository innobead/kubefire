#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail
set -o xtrace

TMP_DIR=/tmp/kubefire
GOARCH=$(go env GOARCH 2>/dev/null)
GOBIN=$(go env GOPATH 2>/dev/null)/bin

SKUBA_VERSION=${SKUBA_VERSION:-}

rm -rf $TMP_DIR && mkdir -p $TMP_DIR
pushd $TMP_DIR

function cleanup() {
  rm -rf $TMP_DIR || true
  popd
}

trap cleanup EXIT ERR INT TERM

function install_skuba() {
  git clone --branch $SKUBA_VERSION https://github.com/SUSE/skuba

  cd skuba
  # remove the br_netfilter check, because it's builtin in the used kernel
  sed -i '/"br_netfilter",/d' ./internal/pkg/skuba/deployments/ssh/kernel.go

  make release
  mv $GOBIN/skuba /usr/local/bin/
}

install_skuba
