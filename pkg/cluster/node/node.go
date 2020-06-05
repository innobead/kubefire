package node

import "github.com/innobead/kubefire/pkg/config"

type Error error

type Node struct {
	Name   string
	Config *config.Node
}

type Manager interface {
	Create(node *config.Node) Error
	Delete(name string) Error
	Get(name string) (*Node, Error)
	List() ([]*Node, Error)
}
