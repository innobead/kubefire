package data

import (
	"github.com/innobead/kubefire/pkg/config"
	"strings"
)

type Cluster struct {
	Name  string
	Spec  config.Cluster
	Nodes []*Node
}

type Node struct {
	Name   string
	Spec   config.Node
	Status NodeStatus
}

type NodeStatus struct {
	Running     bool
	IPAddresses string
	Image       string
	Kernel      string
}

func (n Node) IsMaster() bool {
	return strings.Contains(n.Name, "master")
}
