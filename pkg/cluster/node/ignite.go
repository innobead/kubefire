package node

import (
	"github.com/innobead/kubefire/pkg/config"
)

type IgniteNodeManager struct {
	//TODO use ignite to manage node
}

func (i *IgniteNodeManager) Create(node *config.Node) Error {
	panic("implement me")
}

func (i *IgniteNodeManager) Delete(name string) Error {
	panic("implement me")
}

func (i *IgniteNodeManager) Get(name string) (*Node, Error) {
	panic("implement me")
}

func (i *IgniteNodeManager) List() ([]*Node, Error) {
	panic("implement me")
}
