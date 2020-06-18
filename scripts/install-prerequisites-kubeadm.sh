#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail
set -o xtrace

TMP_DIR=/tmp/kubefire
KUBE_VERSION="v1.18.4" # https://dl.k8s.io/release/stable.txt
KUBE_RELEASE_VERSION="v0.2.7"
CONTAINERD_VERSION="v1.3.4"
CNI_VERSION="v0.8.2"
CRICTL_VERSION="v1.17.0"
RUNC_VERSION="v1.0.0-rc90"

mkdir -p $TMP_DIR
pushd $TMP_DIR

function install_kubeadm() {
  curl -sfOL --remote-name-all https://storage.googleapis.com/kubernetes-release/release/${KUBE_VERSION}/bin/linux/amd64/{kubeadm,kubelet,kubectl}
  chmod +x {kubeadm,kubelet,kubectl}
  mv {kubeadm,kubelet,kubectl} /usr/local/bin/

  curl -sSL "https://raw.githubusercontent.com/kubernetes/release/${KUBE_RELEASE_VERSION}/cmd/kubepkg/templates/latest/deb/kubelet/lib/systemd/system/kubelet.service" | sed "s:/usr/bin:/usr/local/bin:g" >/etc/systemd/system/kubelet.service
  mkdir -p /etc/systemd/system/kubelet.service.d
  curl -sSL "https://raw.githubusercontent.com/kubernetes/release/${KUBE_RELEASE_VERSION}/cmd/kubepkg/templates/latest/deb/kubeadm/10-kubeadm.conf" | sed "s:/usr/bin:/usr/local/bin:g" >/etc/systemd/system/kubelet.service.d/10-kubeadm.conf
  systemctl enable --now kubelet
}

function install_cri_runtime() {
  local version="${CONTAINERD_VERSION:1}"
  local dir=containerd-$version

  curl -sSLO "https://github.com/containerd/containerd/releases/download/$CONTAINERD_VERSION/containerd-$version.linux-amd64.tar.gz"
  mkdir -p $dir
  tar -zxvf $dir*.tar.gz -C $dir
  chmod +x $dir/bin/*
  mv $dir/bin/* /usr/local/bin/

  curl -sSL "https://raw.githubusercontent.com/containerd/containerd/$CONTAINERD_VERSION/containerd.service" >/etc/systemd/system/containerd.service
  mkdir -p /etc/containerd
  containerd config default >/etc/containerd/config.toml
  systemctl enable --now containerd
}

function install_runc() {
    curl -sSL "https://github.com/opencontainers/runc/releases/download/$RUNC_VERSION/runc.amd64" -o runc
    chmod +x runc
    mv runc /usr/local/bin/
}

function install_cni() {
  mkdir -p /opt/cni/bin
  curl -L "https://github.com/containernetworking/plugins/releases/download/${CNI_VERSION}/cni-plugins-linux-amd64-${CNI_VERSION}.tgz" | tar -C /opt/cni/bin -xz
}

function install_kubelet_cri() {
  curl -L "https://github.com/kubernetes-sigs/cri-tools/releases/download/${CRICTL_VERSION}/crictl-${CRICTL_VERSION}-linux-amd64.tar.gz" | tar -C /usr/local/bin -xz
  echo "export CONTAINER_RUNTIME_ENDPOINT=unix:///run/containerd/containerd.sock" >>/etc/profile
}

function cleanup() {
  rm -rf $TMP_DIR || true
}

trap cleanup EXIT ERR INT TERM

install_cni
install_runc
install_kubelet_cri
install_cri_runtime
install_kubeadm

popd
