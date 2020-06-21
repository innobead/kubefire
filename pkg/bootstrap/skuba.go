package bootstrap

import (
	"github.com/innobead/kubefire/pkg/data"
	"github.com/innobead/kubefire/pkg/node"
)

type SkubaBootstrapper struct {
}

func NewSkubaBootstrapper(nodeManager node.Manager) *SkubaBootstrapper {
	return &SkubaBootstrapper{}
}

func (s *SkubaBootstrapper) Deploy(cluster *data.Cluster) error {
	panic("implement me")
}
