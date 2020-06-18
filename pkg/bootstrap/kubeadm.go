package bootstrap

import (
	"github.com/innobead/kubefire/pkg/data"
)

type KubeadmBootstrapper struct {
}

func NewKubeadmBootstrapper() *KubeadmBootstrapper {
	return &KubeadmBootstrapper{}
}

func (k *KubeadmBootstrapper) Deploy(cluster *data.Cluster) error {
	panic("implement me")
}

func (k *KubeadmBootstrapper) init(cluster *data.Cluster) error {
	//TODO a shell script to download, then execute locally to do the all installation stuff...
	// - validate all nodes reachable
	// - look for or download the supported (version) kubeadm, kubelet, kubectl (install into ~/.kubefire/bin/kubeadm)
	// - look for or download the supported CRI runtime manager (cri-o, containerd, docker)
	// - download the shell script to execute

	panic("implement me")
}

func (k *KubeadmBootstrapper) bootstrap(node *data.Node) error {

	//# TODO -- runtime in the code
	//# /etc/sysctl.conf, net.ipv4.ip_forward = 1
	//# sysctl -w net.ipv4.ip_forward=1
	//# echo host >> /etc/hosts
	//# export CONTAINER_RUNTIME_ENDPOINT=unix:///run/containerd/containerd.sock
	//# kubeadm init phase preflight
	//# kubeadm init -v 10
	//# kubectl create -f https://raw.githubusercontent.com/cilium/cilium/v1.8/install/kubernetes/quick-install.yaml
	//# kubectl apply -f https://docs.projectcalico.org/v3.14/manifests/calico.yaml

	panic("implement me")
}

func (k *KubeadmBootstrapper) join(node *data.Node) error {
	panic("implement me")
}
