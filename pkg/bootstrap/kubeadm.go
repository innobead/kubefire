package bootstrap

import (
	"github.com/innobead/kubefire/pkg/data"
)

type KubeadmBootstrapper struct {
}

func NewKubeadmBootstrapper() *KubeadmBootstrapper {
	return &KubeadmBootstrapper{}
}

func (k *KubeadmBootstrapper) Init(cluster *data.Cluster) Error {
	panic("implement me")
}

func (k *KubeadmBootstrapper) Bootstrap(node *data.Node) Error {
	panic("implement me")
}

func (k *KubeadmBootstrapper) Join(node *data.Node) Error {
	panic("implement me")
}

func (k *KubeadmBootstrapper) Install() Error {
	panic("implement me")
}

func (k *KubeadmBootstrapper) Check() Error {
	panic("implement me")
}
