package node

import (
	"github.com/innobead/kubefire/pkg/config"
	"github.com/innobead/kubefire/pkg/data"
)

type Error error

type Type string

const (
	Admin  Type = "admin"
	Master Type = "master"
	Worker Type = "worker"
)

type Manager interface {
	CreateNodes(nodeType Type, node *config.Node) Error
	DeleteNodes(nodeType Type, node *config.Node) Error
	DeleteNode(name string) Error
	GetNode(name string) (*data.Node, Error)
	ListNodes(clusterName string) ([]*data.Node, Error)
}
