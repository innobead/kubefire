package cluster

import (
	"github.com/innobead/kubefire/pkg/config"
)

type Error error

type Cluster struct {
	Name   string
	Config *config.Cluster
}

type Manager interface {
	Init(cluster *config.Cluster) Error
	Create(cluster *config.Cluster) Error
	Delete(cluster *config.Cluster) Error
	Get(name string) (*Cluster, Error)
	List() ([]*Cluster, Error)
}

type DefaultManager struct {
	// TODO use IgniteNodeManager
}

func (d *DefaultManager) Init(cluster *config.Cluster) Error {
	panic("implement me")
}

func (d *DefaultManager) Create(cluster *config.Cluster) Error {
	panic("implement me")
}

func (d *DefaultManager) Delete(cluster *config.Cluster) Error {
	panic("implement me")
}

func (d *DefaultManager) Get(name string) (*Cluster, Error) {
	panic("implement me")
}

func (d *DefaultManager) List() ([]*Cluster, Error) {
	panic("implement me")
}
