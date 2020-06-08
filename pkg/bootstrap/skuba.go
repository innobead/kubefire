package bootstrap

import (
	"github.com/innobead/kubefire/pkg/data"
)

type SkubaBootstrapper struct {
}

func NewSkubaBootstrapper() *SkubaBootstrapper {
	return &SkubaBootstrapper{}
}

func (s *SkubaBootstrapper) Init(cluster *data.Cluster) Error {
	panic("implement me")
}

func (s *SkubaBootstrapper) Bootstrap(node *data.Node) Error {
	panic("implement me")
}

func (s *SkubaBootstrapper) Join(node *data.Node) Error {
	panic("implement me")
}

func (s *SkubaBootstrapper) Install() Error {
	panic("implement me")
}

func (k *SkubaBootstrapper) Check() Error {
	panic("implement me")
}
