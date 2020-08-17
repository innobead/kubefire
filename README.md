# What is KubeFire?

KubeFire is to manage Kubernetes clusters running on FireCracker microVMs via **weaveworks/ignite**. 

- No need to have KVM qocws image for rootfs and kernel. Ignite uses independent rootfs and kernel from OCI images.
- Ignite uses container managment engine like docker or containerd to manage Firecracker processes running in containers.
- Have different bootstappers to provision Kubernetes clusters like Kubeadm, K3s, and SUSE Skuba. 

# Getting Started

## Installing KubeFire

There is no official release, so please make sure go 1.14 installed, then build and install `kubefire` in the `GOBIN` path.

```
make install
```

## Installing Prerequisites

To be able to run kubefire commands w/o issues like node/cluster management, there are some prerequisites to have. Please run `kubefire install` command with root permission (or sudo without password) to have these prerequisites via the below steps.

- Check virtualization supported
- Install necessary components including runc, containerd, CNI plugins, and Ignite

> Note: 
> - To uninstall the prerequisites, run `kubefire uninstall`.
> - To check the installation status, run `kubefire info`. 

[![asciicast](https://asciinema.org/a/tQKqYjojnsgZOjZqrGbF9Zqh0.svg)](https://asciinema.org/a/tQKqYjojnsgZOjZqrGbF9Zqh0)

## Bootstrapping Cluster

### Kubeadm (K8s 1.18.8)

```
kubefire cluster create --bootstrapper=kubeadm demo
```

[![asciicast](https://asciinema.org/a/lQfFfMa1zCXWvz321eUqhNyxB.svg)](https://asciinema.org/a/lQfFfMa1zCXWvz321eUqhNyxB)

### K3s (K8s 1.18.8)

Please note that K3s only officially supports Ubuntu 16.04 and 18.04, the kernel versions of which are 4.4 and 4.15. 
Therefore, if using the prebuilt kernels, please use `4.19` (which is the default kernel used) instead of `5.4`, otherwise there will be some unexpected errors happening. 
For rootfs, it's no problem to use other non-Ubuntu images.

```
kubefire cluster create demo --bootstrapper=k3s
```

#### Add extra K3s installation options

To add extra installation options of the server or agent nodes, use `--extra-options` of `cluster create` command to provide `ServerOpts` or `AgentOpts` key-value pairs as the below example. 

> Note: the key-value pairs in `--extra-options` are separated by comma.

- Add any options of `k3s server` into `ServerOpts='<k3s server option1>, <k3s server option2>, ...'`.
- Add any options of `k3s agent` into `AgentOpts='<k3s agent option1>, <k3s agent option2>, ...'`.

```
kubefire cluster create demo-k3s --bootstrapper k3s --extra-opts="ServerOpts='--disable=traefik --disable=metrics-server'"
```

[![asciicast](https://asciinema.org/a/HqmfS4wZP7pPVS3E7M7gwAzmA.svg)](https://asciinema.org/a/HqmfS4wZP7pPVS3E7M7gwAzmA)

### SUSE Skuba (K8s 1.17.9)

```
kubefire cluster create demo --bootstrapper=skuba --extra-opts="RegisterCode=<Product Register Code>"
```

## Accessing Cluster

During bootstrapping, the cluster folder is created at `~/.kubefire/clusters/<cluster name>`. After bootstrapping, there are several files generated in the folder.

- **admin.conf**
  
  The kubeconfig, downloaded from one of master nodes

- **cluster.yaml**

  The cluster config manifest is for creating the cluster. There is no declarative management based on it for now, but maybe it will be introduced in the future.

- **key, key.pub**
  
  The private and public keys for SSH authentication to all nodes in the cluster.
  
There are two ways below to operate the deployed cluster. After having a valid KUBECONFIG setup, run kubectl commands as usual.

1. run `eval $(kubefire cluster env <cluster name>)` to update KUBECONFIG pointing to `~/.kubefire/clusters/<cluster name>/admin.conf`.
2. run `kubefire node ssh <master node name>` to ssh to one of master nodes, then update KUBECONFIG pointing to `/etc/kubernetes/admin.conf`. For K3s, the kubeconfig is `/etc/rancher/k3s/k3s.yaml` instead.

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

# Show runtime information
kubefire info

# Install prerequisites
kubefire install 

# Uninstall prerequisites
kubefire uninstall

# Create a cluster
kubefire cluster create

# Delete clusters
kubefire cluster delete

# Get a cluster info
kubefire cluster get

# List clusters
kubefire cluster list

# Download cluster kubeconfig
kubefire cluster download

# Print environment variables of cluster (ex: KUBECONFIG)
kubefire cluster env

# SSH to a node
kubefire node ssh
```
 
# Supported Container Images for RootFS and Kernel

Besides below prebuilt images, you can also use the images provided by [weaveworks/ignite](https://github.com/weaveworks/ignite/tree/master/images).

## RootFS images
- docker.io/innobead/kubefire-opensuse-leap:15.1, 15.2
- docker.io/innobead/kubefire-sle15:15.1, 15.2
- docker.io/innobead/kubefire-centos:8
- docker.io/innobead/kubefire-ubuntu:18.04
- docker.io/innobead/kubefire-ubuntu:20.10

## Kernel images (w/ AppArmor enabled)
- docker.io/innobead/kubefire-kernel-5.4.43-amd64:latest
- docker.io/innobead/kubefire-kernel-4.19.125-amd64:latest (default)

## References

- [Firecracker](https://github.com/firecracker-microvm/firecracker)
- [Ignite](https://github.com/weaveworks/ignite)
- [K3s](https://github.com/rancher/k3s) 

