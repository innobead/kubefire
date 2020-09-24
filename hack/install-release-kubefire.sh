#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail
set -o xtrace

function get_latest_release() {
  curl -sfSL "https://api.github.com/repos/innobead/kubefire/releases/latest" |
    grep '"tag_name":' |
    sed -E 's/.*"([^"]+)".*/\1/'
}

filename="kubefire-linux-amd64"
if [[ "$(uname -m)" == "aarch64" ]]; then
  filename="kubefire-linux-arm64"
fi

echo $filename
# shellcheck disable=SC2046
curl -sfSLO "https://github.com/innobead/kubefire/releases/download/$(get_latest_release)/$filename" && chmod +x $filename && sudo mv $filename /usr/local/bin/kubefire
