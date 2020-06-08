package data

import "github.com/innobead/kubefire/pkg/config"

type Cluster struct {
	Name   string
	Config *config.Cluster
}

type Node struct {
	Name   string
	Config *config.Node
}
