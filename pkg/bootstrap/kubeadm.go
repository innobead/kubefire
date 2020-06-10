package bootstrap

import (
	"github.com/innobead/kubefire/pkg/data"
)

type KubeadmBootstrapper struct {
}

func NewKubeadmBootstrapper() *KubeadmBootstrapper {
	return &KubeadmBootstrapper{}
}

func (k *KubeadmBootstrapper) Init(cluster *data.Cluster) error {
	panic("implement me")
}

func (k *KubeadmBootstrapper) Bootstrap(node *data.Node) error {
	panic("implement me")
}

func (k *KubeadmBootstrapper) Join(node *data.Node) error {
	panic("implement me")
}

func (k *KubeadmBootstrapper) InstallRequirements() error {
	panic("implement me")
}

func (k *KubeadmBootstrapper) CheckRequirements() error {
	panic("implement me")
}
