package bootstrap

import (
	"github.com/innobead/kubefire/pkg/data"
	"github.com/innobead/kubefire/pkg/node"
)

type SkubaBootstrapper struct {
	nodeManager node.Manager
}

func NewSkubaBootstrapper(nodeManager node.Manager) *SkubaBootstrapper {
	return &SkubaBootstrapper{
		nodeManager: nodeManager,
	}
}

func (s *SkubaBootstrapper) Deploy(cluster *data.Cluster) error {
	panic("implement me")
}
