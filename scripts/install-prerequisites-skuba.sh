#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail
set -o xtrace

TMP_DIR=/tmp/kubefire
GOARCH=$(go env GOARCH 2>/dev/null || echo "amd64")
GOBIN=$(go env GOBIN || echo "/usr/local/bin")

mkdir -p $TMP_DIR
pushd $TMP_DIR

function cleanup() {
  rm -rf $TMP_DIR || true
  popd
}

trap cleanup EXIT ERR INT TERM

function install_skuba() {
  git clone --branch v1.3.5 https://github.com/SUSE/skuba
  cd skuba
  make release
  mv $GOBIN/skuba ~/.kubefire/bin/
}

install_skuba
