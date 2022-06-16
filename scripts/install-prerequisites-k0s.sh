#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail
set -o xtrace

TMP_DIR=/tmp/kubefire

K0S_VERSION=${K0S_VERSION:-}
K0S_CONFIG=${K0S_CONFIG:-}
K0S_CMD_OPTS=${K0S_CMD_OPTS:-}
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

  if [[ $(command -v apt-get) ]]; then
    apt update
    apt install ipip
  fi
  modprobe ipip

  curl -sfSL "$url" -o k0s
  chmod +x k0s && sudo mv k0s /usr/local/bin/
}

function create_controller() {
  create_config
  enable_service
}

function join_node() {
  enable_service
}

function create_config() {
  mkdir -p /etc/k0s || true
  if [ -n "$K0S_CONFIG" ]; then
    echo "$K0S_CONFIG" >/etc/k0s/config.yaml
  else
    k0s config create >/etc/k0s/config.yaml || k0s default-config >/etc/k0s/config.yaml
  fi
}

function enable_service() {
  if k0s install -h >/dev/null; then
    # shellcheck disable=SC2086
    k0s install $K0S_CMD_OPTS
    k0s start
  else
    # deprecated
    cat <<EOF >/etc/systemd/system/k0s.service
[Unit]
Description=K0s server
Wants=network-online.target
After=network-online.target

[Install]
WantedBy=multi-user.target

[Service]
KillMode=process
Delegate=yes
LimitNOFILE=1048576
LimitNPROC=infinity
LimitCORE=infinity
TasksMax=infinity
TimeoutStartSec=0
Restart=always
RestartSec=5s
ExecStartPre=-/sbin/modprobe ipip
ExecStart=/usr/local/bin/k0s $K0S_START_CMD $K0S_CMD_OPTS
EOF

    systemctl enable k0s.service
    systemctl start k0s.service
  fi
}

$1
