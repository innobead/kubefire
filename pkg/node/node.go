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
	Delete(name string) Error
	Get(name string) (*data.Node, Error)
	List(clusterName string) ([]*data.Node, Error)
}
