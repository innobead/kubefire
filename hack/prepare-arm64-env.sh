#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail
set -o xtrace

[[ $(uname -m) != "aarh64" ]] && echo "It's not ARM64 env!" > /dev/stderr && exit 1

echo "Installing packages ..."
sudo apt update
sudo apt install unzip
sudo apt install cpu-checker
sudo apt install qemu-kvm libvirt-daemon-system libvirt-clients bridge-utils virtinst virt-manager
sudo apt install make
sudo apt install build-essential docker.io
sudo apt install btrfs-progs libbtrfs-dev pkg-config libseccomp-dev

sudo snap install --classic go
sudo snap install --classic kubectl

echo "Building containerd ARM64 artifacts ..."
go get github.com/containerd/containerd
cd "$(go env GOPATH)/src/github.com/containerd/containerd"
git checkout -b v1.4.3 v1.4.3
make
sudo make install

echo "Building runc ARM64 artifacts ..."
go get github.com/opencontainers/runc
cd "$(go env GOPATH)/src/github.com/opencontainers/runc"
make
sudo make install

echo "Installing Kubefire prerequisites ..."
curl -sfSL https://raw.githubusercontent.com/innobead/kubefire/master/hack/install-release-kubefire.sh | bash
kubefire install
kubefire info
