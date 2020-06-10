package data

import "github.com/innobead/kubefire/pkg/config"

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
	Running bool
}
