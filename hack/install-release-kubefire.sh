#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail
set -o xtrace

function get_latest_release() {
  curl -sSL "https://api.github.com/repos/innobead/kubefire/releases/latest" |
    grep '"tag_name":' |
    sed -E 's/.*"([^"]+)".*/\1/'
}

curl -sSLO "https://github.com/innobead/kubefire/releases/download/$(get_latest_release)/kubefire" && chmod +x kubefire && sudo mv kubefire /usr/local/bin
