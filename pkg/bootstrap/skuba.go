package bootstrap

import (
	"github.com/innobead/kubefire/pkg/data"
)

type SkubaBootstrapper struct {
}

func NewSkubaBootstrapper() *SkubaBootstrapper {
	return &SkubaBootstrapper{}
}

func (s *SkubaBootstrapper) init(cluster *data.Cluster) error {
	panic("implement me")
}

func (s *SkubaBootstrapper) bootstrap(node *data.Node) error {
	panic("implement me")
}

func (s *SkubaBootstrapper) Deploy(cluster *data.Cluster) error {
	panic("implement me")
}

func (s *SkubaBootstrapper) join(node *data.Node) error {
	panic("implement me")
}
