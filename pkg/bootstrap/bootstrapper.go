package bootstrap

import (
	"github.com/innobead/kubefire/pkg/cluster"
	"github.com/innobead/kubefire/pkg/cluster/node"
)

type Error error

const (
	KUBEADM = "kubeadm"
	SKUBA   = "skuba"
)

var BuiltinTypes = map[string]func() Bootstrapper{
	KUBEADM: NewKubeadmBootstrapper,
	SKUBA:   NewSkubaBootstrapper,
}

type Bootstrapper interface {
	Init(cluster *cluster.Cluster) Error
	Bootstrap(node *node.Node) Error
	Join(node *node.Node) Error
}

type BootstrapperInstaller interface {
	Install() Error
	Check() Error
}
