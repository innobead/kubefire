# What is KubeFire?

KubeFire is to manage Kubernetes clusters running on FireCracker microVMs via **weaveworks/ignite**. 

- No need to have KVM qocws image for rootfs and kernel. Ignite uses independent rootfs and kernel from OCI images.
- Ignite uses container managment engine like docker or containerd to manage Firecracker processes running in containers.
- Have different bootstappers to provision Kubernetes clusters like kubeadm and SUSE skuba. 

# Getting Started

## Installing KubeFire

There is no official release, so please make sure go 1.14 installed, then build and install kubefire in the `GOBIN` path.

```
make install
```

## Installing Prerequisites

Run the below command with root permission/sudo without password for below things.

- Check virtualization supported
- Install necessary components including runc, containerd, CNI plugins, and Ignite

```
kubefire install
```

## Bootstrapping Cluster

### Kubeadm (K8s 1.18.5)

```
kubefire cluster create --bootstrapper=kubeadm demo
```

[![asciicast](https://asciinema.org/a/344343.svg)](https://asciinema.org/a/344343)

### SUSE Skuba (K8s 1.17.4)

```
kubefire cluster create demo --bootstrapper=skuba --extra-opts="RegisterCode=<Product Register Code>"
```

# Usage

## CLI Commands

Make sure to run kubefire commands with root permission or sudo without password, because ignite needs root permission to manage Firecracker VMs for now, but it is planned to improve in the future release.

```
KubeFire, manage Kubernetes clusters on FireCracker microVMs

Usage:
  kubefire [flags]
  kubefire [command]

Available Commands:
  cluster     Manage cluster
  help        Help about any command
  install     Install prerequisites
  node        Manage node
  uninstall   Uninstall prerequisites
  version     Show version

Flags:
  -h, --help               help for kubefire
      --log-level string   log level, options: [panic, fatal, error, warning, info, debug, trace] (default "info")
      --output string      output format, options: [default, json, yaml] (default "default")

Use "kubefire [command] --help" for more information about a command.

```

```
# Show version
kubefire version

# Install necessary components for cluster management
kubefire install 

# Uninstall ncessary components to clean up the environment
kubefire uninstall

# Create a cluster
kubefire cluster create

# Delete a cluster
kubefire cluster delete

# Get a cluster info
kubefire cluster get

# List clusters
kubefire cluster list

# Download cluster kubeconfig
kubefire cluster download

# SSH to a node
kubefire node ssh
```
 
# Supported Container Images for RootFS and Kernel

Besides below prebuilt images, you can also use the images provided by [weaveworks/ignite](https://github.com/weaveworks/ignite/tree/master/images).

## RootFS images
- docker.io/innobead/kubefire-opensuse-leap:15.1
- docker.io/innobead/kubefire-sle15:15.1
- docker.io/innobead/kubefire-centos:8
- docker.io/innobead/kubefire-ubuntu:20.10

## Kernel images (w/ AppArmor enabled)
- docker.io/innobead/kubefire-kernel-5.4.43-amd64:latest
- docker.io/innobead/kubefire-kernel-4.19.125-amd64:latest

## References

- [Firecracker](https://github.com/firecracker-microvm/firecracker)
- [Ignite](https://github.com/weaveworks/ignite)
 
