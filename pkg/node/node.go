package node

import (
	"github.com/innobead/kubefire/pkg/config"
	"github.com/innobead/kubefire/pkg/data"
)

type Type string

const (
	Admin  Type = "admin"
	Master Type = "master"
	Worker Type = "worker"
)

type Manager interface {
	CreateNodes(nodeType Type, node *config.Node) error
	DeleteNodes(nodeType Type, node *config.Node) error
	DeleteNode(name string) error
	GetNode(name string) (*data.Node, error)
	ListNodes(clusterName string) ([]*data.Node, error)
}
