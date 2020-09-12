# What is KubeFire?:fire: 

KubeFire is to create and manage Kubernetes clusters running on FireCracker microVMs via **weaveworks/ignite**. 

- No need to have KVM qocws image for rootfs and kernel. Ignite uses independent rootfs and kernel from OCI images.
- Ignite uses container managment engine like docker or containerd to manage Firecracker processes running in containers.
- Have different bootstappers to provision Kubernetes clusters like Kubeadm, K3s, and SUSE Skuba. 

# Getting Started

## Installing KubeFire

For official releases, please install the latest release as below command.

```console
curl -sSL https://raw.githubusercontent.com/innobead/kubefire/master/hack/install-release-kubefire.sh | sh -
```

For development purpose, please make sure go 1.14 installed, then build and install `kubefire` in the `GOBIN` path.

```console
make install
```

## Quickstart

Running below commands is to quickly have a cluster deployed by kubeadm running in minutes.

```console
kubefire install
kubefire cluster create demo
```

## Installing or Update Prerequisites

To be able to run kubefire commands w/o issues like node/cluster management, there are some prerequisites to have. 
Please run `kubefire install` command with root permission (or sudo without password) to install or update these prerequisites via the below steps.

- Check virtualization supported
- Install necessary components including runc, containerd, CNI plugins, and Ignite

> Note: 
> - To uninstall the prerequisites, run `kubefire uninstall`.
> - To check the installation status, run `kubefire info`. 

[![asciicast](https://asciinema.org/a/tQKqYjojnsgZOjZqrGbF9Zqh0.svg)](https://asciinema.org/a/tQKqYjojnsgZOjZqrGbF9Zqh0)

## Bootstrapping Cluster

### Bootstrap with selectable Kubernetes versions

From v0.2.0, Kubefire supports user to create a cluster with a specific version supported by built-in bootstrappers in the below cases.

```console
# Create a cluster with the latest versions w/o any specified version
kubefire cluster create demo

# Create a cluster with the latest patch version of v1.18
kubefire cluster create demo --version=v1.18

# Create a cluster with a valid specific version v1.18.8
kubefire cluster create demo --version=v1.18.8

# Create a cluster with the latest patch version of supported minor releases
kubefire cluster create demo --version=v1.17
kubefire cluster create demo --version=v1.16

# If the version is outside the supported versions (last 3 minor versions given the latest is v1.18), the cluster creation will be not supported 
kubefire cluster create demo --version=v1.15
```

### Kubeadm
> Supports [the latest supported version](https://dl.k8s.io/release/stable.txt) and last 3 minor versions.

```console
kubefire cluster create demo --bootstrapper=kubeadm
```

[![asciicast](https://asciinema.org/a/lQfFfMa1zCXWvz321eUqhNyxB.svg)](https://asciinema.org/a/lQfFfMa1zCXWvz321eUqhNyxB)

### K3s
> Supports [the latest supported version](https://update.k3s.io/v1-release/channels/latest) and last 3 minor versions.

Please note that K3s only officially supports Ubuntu 16.04 and 18.04, the kernel versions of which are 4.4 and 4.15. 
Therefore, if using the prebuilt kernels, please use `4.19` (which is the default kernel used) instead of `5.4`, otherwise there will be some unexpected errors happening. 
For rootfs, it's no problem to use other non-Ubuntu images.

```console
kubefire cluster create demo --bootstrapper=k3s
```

#### Add extra K3s installation options

To add extra installation options of the server or agent nodes, use `--extra-options` of `cluster create` command to provide `ServerOpts` or `AgentOpts` key-value pairs as the below example. 

> Note: the key-value pairs in `--extra-options` are separated by comma.

- Add any options of `k3s server` into `ServerOpts='<k3s server option1>, <k3s server option2>, ...'`.
- Add any options of `k3s agent` into `AgentOpts='<k3s agent option1>, <k3s agent option2>, ...'`.

```console
kubefire cluster create demo --bootstrapper=k3s --extra-opts="ServerOpts='--disable=traefik --disable=metrics-server'"
```

[![asciicast](https://asciinema.org/a/HqmfS4wZP7pPVS3E7M7gwAzmA.svg)](https://asciinema.org/a/HqmfS4wZP7pPVS3E7M7gwAzmA)

### SUSE Skuba (K8s 1.17.9)

```console
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

```console
KubeFire, creates and manages Kubernetes clusters using FireCracker microVMs

Usage:
  kubefire [flags]
  kubefire [command]

Available Commands:
  cluster     Manage clusters
  help        Help about any command
  info        Show info of prerequisites, supported K8s/K3s versions
  install     Install prerequisites
  kubeconfig  Manage kubeconfig of clusters
  node        Manage nodes
  uninstall   Uninstall prerequisites
  version     Show version

Flags:
  -h, --help               help for kubefire
      --log-level string   log level, options: [panic, fatal, error, warning, info, debug, trace] (default "info")
  -o, --output string      output format, options: [default, json, yaml] (default "default")

Use "kubefire [command] --help" for more information about a command.
```

```console
# Show version
kubefire version

# Show prerequisites information
kubefire info

# Show supported K8s/K3s versions by builtin bootstrappers
kubefire info -b

# Install or Update prerequisites
kubefire install 

# Uninstall prerequisites
kubefire uninstall

# Create a cluster
kubefire cluster create

# Create a cluster w/ a selected version
kubefire cluster create --version=[v<MAJOR>.<MINOR>.<PATCH> | v<MAJOR>.<MINOR>]

# Delete clusters
kubefire cluster delete

# Get a cluster info
kubefire cluster get

# Stop a cluster
kubefire cluster stop

# Start a cluster
kubefire cluster start

# Restart a cluster
kubefire cluster restart

# List clusters
kubefire cluster list

# Print environment variables of cluster (ex: KUBECONFIG)
kubefire cluster env

# Print cluster kubeconfig
kubefire kubeconfig get

# Download cluster kubeconfig
kubefire kubeconfig download

# SSH to a node
kubefire node ssh

# Get a node info
kubefire node get

# Stop a node
kubefire node stop

# Start a node
kubefire node start

# Restart a node
kubefire node restart
```

# Troubleshooting

If encountering any unexpected behavior like ignite can't allocate valid IPs to the created VMs. 
Please try to clean up the environment, then verify again. If the issues still cannot be resolved by environment cleanup, please help create issues. 

```console
kubefire unisntall
kubefire install
```  
 
# Supported Container Images for RootFS and Kernel

Besides below prebuilt images, you can also use the images provided by [weaveworks/ignite](https://github.com/weaveworks/ignite/tree/master/images).

## RootFS images
- ghcr.io/innobead/kubefire-opensuse-leap:15.1, 15.2
- ghcr.io/innobead/kubefire-sle15:15.1, 15.2
- ghcr.io/innobead/kubefire-centos:8
- ghcr.io/innobead/kubefire-ubuntu:18.04
- ghcr.io/innobead/kubefire-ubuntu:20.10

## Kernel images (w/ AppArmor enabled)
- ghcr.io/innobead/kubefire-ignite-kernel:5.4.43-amd64
- ghcr.io/innobead/kubefire-ignite-kernel:4.19.125-amd64 (default)

## References

- [Firecracker](https://github.com/firecracker-microvm/firecracker)
- [Ignite](https://github.com/weaveworks/ignite)
- [K3s](https://github.com/rancher/k3s) 

