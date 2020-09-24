#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail
#set -o xtrace

PROJECT_DIR=$(readlink -f "$(dirname "${BASH_SOURCE[0]}")/..")
PROJECT=$(basename "$PROJECT_DIR")
IMAGES="centos:8 ubuntu:18.04 ubuntu:20.04 ubuntu:20.10 opensuse-leap:15.1 opensuse-leap:15.2"
KERNELS=$(ls "${PROJECT_DIR}/build/kernels" | grep -v "README.md" | sed -E 's/config-(arm64|amd64)-//;' | awk '!a[$0]++')
GENERATED_DIR=${PROJECT_DIR}/generated
IMAGE_LIST_FILE=${GENERATED_DIR}/image.list
KERNEL_LIST_FILE=${GENERATED_DIR}/kernel.list
ARCH_LIST="amd64 arm64"
CR_IMAGE_PREFIX=ghcr.io/innobead

function rootfs_image_urls() {
  for img in $IMAGES; do
    echo "${CR_IMAGE_PREFIX}/${PROJECT}-${img}"
  done
}

function kernel_image_urls() {
  for arch in $ARCH_LIST; do
    for img in $KERNELS; do
      echo "${CR_IMAGE_PREFIX}/${PROJECT}-ignite-kernel:${img}-${arch}"
    done
  done
}

function generate_image_lists() {
  mkdir -p "${GENERATED_DIR}" || true

  rootfs_image_urls > "$IMAGE_LIST_FILE"
  kernel_image_urls > "$KERNEL_LIST_FILE"
}

if [[ $# -ne 1 ]]; then
  echo "Please specify one argument. (choices: --generate, --image, --kernel)" >/dev/stderr
  exit 1
fi

# shellcheck disable=SC2086
case $1 in
--generate)
  generate_image_lists
  ;;

--image)
  echo $IMAGES
  ;;

--kernel)
  echo $KERNELS
  ;;
esac
