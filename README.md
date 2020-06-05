# Introduction

KubeFire, deploy Kubernetes clusters in seconds and manage the cluster lifecycle by Docker or Firecracker.

# Prerequisites

- Docker or ContainerD
- CNI plugins

# Usage

```
kubefire install

kubefire cluster init <name> --master 1 --worker 2 --provider [kubeadm|skuba]
kubefire cluster create <name>
kubefire cluster deploy <name>
kubefire cluster info <name>
kubefire cluster destory <name>
kubefire cluster delete <name>
kubefire cluster upgrade <name>

kubefire node info
```

# Design 

## `image build`

Build an OCI compatible image for root file system usage by Firecracker MicroVM. 

## `cluster init`

Create a cluster manifest by using `ignite`. Besides, you can customize the settings and provisions of cluster and nodes by extensible shell scripts.

## `cluster create`

Use `ignite` to create defined numbers of nodes.

## `cluster deploy`

Deploy the Kubernetes cluster by using a specified provisioner like kubeadm, skuba (SUSE solution), etc.

## `cluster info`

Show the cluster info like cluster configuration, node info, etc

## `cluster destroy`

Destroy the cluster but optionally leave the nodes

## `cluster delete` 

Delete the nodes, and the cluster

## `cluster upgrade`

Upgrade the cluster by using the specified provisioner

## `node info`

Show the detailed info of a node

# Limitations

- The usage depends on the prerequisites of `footloose`, `ignite`, and `firecracker`.