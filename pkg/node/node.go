package node

import (
	"github.com/innobead/kubefire/pkg/config"
	"github.com/innobead/kubefire/pkg/data"
)

type Error error

type Manager interface {
	CreateNodes(clusterName string, node *config.Node) Error
	DeleteNodes(clusterName string, node *config.Node) Error
	Delete(name string) Error
	Get(name string) (*data.Node, Error)
	List() ([]*data.Node, Error)
}
