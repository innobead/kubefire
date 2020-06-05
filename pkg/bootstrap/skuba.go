package bootstrap

import (
	"github.com/innobead/kubefire/pkg/cluster"
	"github.com/innobead/kubefire/pkg/cluster/node"
)

type SkubaBootstrapper struct {
}

func NewSkubaBootstrapper() Bootstrapper {
	return &SkubaBootstrapper{}
}

func (s *SkubaBootstrapper) Init(cluster *cluster.Cluster) Error {
	panic("implement me")
}

func (s *SkubaBootstrapper) Bootstrap(node *node.Node) Error {
	panic("implement me")
}

func (s *SkubaBootstrapper) Join(node *node.Node) Error {
	panic("implement me")
}

func (s *SkubaBootstrapper) Install() Error {
	panic("implement me")
}

func (k *SkubaBootstrapper) Check() Error {
	panic("implement me")
}
