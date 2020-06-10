package bootstrap

import (
	"github.com/innobead/kubefire/pkg/data"
)

type SkubaBootstrapper struct {
}

func NewSkubaBootstrapper() *SkubaBootstrapper {
	return &SkubaBootstrapper{}
}

func (s *SkubaBootstrapper) Init(cluster *data.Cluster) error {
	panic("implement me")
}

func (s *SkubaBootstrapper) Bootstrap(node *data.Node) error {
	panic("implement me")
}

func (s *SkubaBootstrapper) Join(node *data.Node) error {
	panic("implement me")
}

func (s *SkubaBootstrapper) InstallRequirements() error {
	panic("implement me")
}

func (k *SkubaBootstrapper) CheckRequirements() error {
	panic("implement me")
}
