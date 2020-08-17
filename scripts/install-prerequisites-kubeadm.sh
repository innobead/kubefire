#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail
set -o xtrace

TMP_DIR=/tmp/kubefire
GOARCH=$(go env GOARCH 2>/dev/null || echo "amd64")

KUBEFIRE_VERSION=${KUBEFIRE_VERSION:-}
KUBE_VERSION=${KUBE_VERSION:-""} # https://dl.k8s.io/release/stable.txt
KUBE_RELEASE_VERSION=${KUBE_RELEASE_VERSION:-"v0.3.4"}
CONTAINERD_VERSION=${CONTAINERD_VERSION:-""}
CNI_VERSION=${CNI_VERSION:-""}
RUNC_VERSION=${RUNC_VERSION:-""}
CRICTL_VERSION=${CRICTL_VERSION:-"v1.18.0"}

if [ -z "$KUBEFIRE_VERSION" ] || [ -z "$KUBE_VERSION" ] || [ -z "$KUBE_RELEASE_VERSION" ] || [ -z "$CONTAINERD_VERSION" ] || [ -z "$IGNITE_VERION" ] || [ -z "$CNI_VERSION" ] || [ -z "$RUNC_VERSION" ]; then
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

function install_kubeadm() {
  curl -sfOL --remote-name-all "https://storage.googleapis.com/kubernetes-release/release/${KUBE_VERSION}/bin/linux/${GOARCH}/{kubeadm,kubelet,kubectl}"
  chmod +x {kubeadm,kubelet,kubectl}
  sudo mv {kubeadm,kubelet,kubectl} /usr/local/bin/

  curl -sSL "https://raw.githubusercontent.com/kubernetes/release/${KUBE_RELEASE_VERSION}/cmd/kubepkg/templates/latest/deb/kubelet/lib/systemd/system/kubelet.service" | sudo sed "s:/usr/bin:/usr/local/bin:g" >/etc/systemd/system/kubelet.service
  mkdir -p /etc/systemd/system/kubelet.service.d
  curl -sSL "https://raw.githubusercontent.com/kubernetes/release/${KUBE_RELEASE_VERSION}/cmd/kubepkg/templates/latest/deb/kubeadm/10-kubeadm.conf" | sudo sed "s:/usr/bin:/usr/local/bin:g" >/etc/systemd/system/kubelet.service.d/10-kubeadm.conf
  sudo systemctl enable --now kubelet
}

function install_containerd() {
  local version="${CONTAINERD_VERSION:1}"
  local dir=containerd-$version

  curl -sSLO "https://github.com/containerd/containerd/releases/download/${CONTAINERD_VERSION}/containerd-${version}-linux-${GOARCH}.tar.gz"
  mkdir -p $dir
  tar -zxvf $dir*.tar.gz -C $dir
  chmod +x $dir/bin/*
  sudo mv $dir/bin/* /usr/local/bin/

  curl -sSLO "https://raw.githubusercontent.com/containerd/containerd/${CONTAINERD_VERSION}/containerd.service"
  sudo mv containerd.service /etc/systemd/system/containerd.service
  sudo mkdir -p /etc/containerd
  containerd config default | sudo tee /etc/containerd/config.toml >/dev/null
  sudo systemctl enable --now containerd
}

function install_runc() {
  curl -sSL "https://github.com/opencontainers/runc/releases/download/${RUNC_VERSION}/runc.${GOARCH}" -o runc
  chmod +x runc
  sudo mv runc /usr/local/bin/
}

function install_cni() {
  mkdir -p /opt/cni/bin
  curl -sSL "https://github.com/containernetworking/plugins/releases/download/${CNI_VERSION}/cni-plugins-linux-${GOARCH}-${CNI_VERSION}.tgz" | tar -C /opt/cni/bin -xz
}

function install_kubelet_cri() {
  curl -sSL "https://github.com/kubernetes-sigs/cri-tools/releases/download/${CRICTL_VERSION}/crictl-${CRICTL_VERSION}-linux-${GOARCH}.tar.gz" | sudo tar -C /usr/local/bin -xz
  echo "export CONTAINER_RUNTIME_ENDPOINT=unix:///run/containerd/containerd.sock" >>/etc/profile
}

install_cni
install_runc
install_kubelet_cri
install_containerd
install_kubeadm

