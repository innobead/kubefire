package bootstrap

import (
	"github.com/innobead/kubefire/pkg/cluster"
	"github.com/innobead/kubefire/pkg/cluster/node"
)

type KubeadmBootstrapper struct {
}

func NewKubeadmBootstrapper() Bootstrapper {
	return &KubeadmBootstrapper{}
}

func (k *KubeadmBootstrapper) Init(cluster *cluster.Cluster) Error {
	panic("implement me")
}

func (k *KubeadmBootstrapper) Bootstrap(node *node.Node) Error {
	panic("implement me")
}

func (k *KubeadmBootstrapper) Join(node *node.Node) Error {
	panic("implement me")
}

func (k *KubeadmBootstrapper) Install() Error {
	panic("implement me")
}

func (k *KubeadmBootstrapper) Check() Error {
	panic("implement me")
}
