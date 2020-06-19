#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail
set -o xtrace

TMP_DIR=/tmp/kubefire
GOARCH=$(go env GOARCH 2>/dev/null || echo "amd64")

mkdir -p $TMP_DIR
pushd $TMP_DIR

function cleanup() {
  rm -rf $TMP_DIR || true
}

trap cleanup EXIT ERR INT TERM
