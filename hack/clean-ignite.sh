#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail
set -o xtrace

sudo ignite rm -f $(sudo ignite ps -aq) &>/dev/null || echo "> No VMs to delete from ignite"
sudo ignite rmi $(sudo ignite images ls | awk '{print $$1}' | sed '1d') &>/dev/null || echo "> No images to delete from ignite"
sudo ignite rmk $(sudo ignite kernels ls | awk '{print $$1}' | sed '1d') &>/dev/null || echo "> No kernels to delete from ignite"
sudo ctr -n firecracker i rm $(sudo ctr -n firecracker images ls | awk '{print $$1}' | sed '1d') &>/dev/null
