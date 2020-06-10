package bootstrap

import (
	"github.com/innobead/kubefire/pkg/data"
)

const (
	KUBEADM = "kubeadm"
	SKUBA   = "skuba"
)

var BuiltinTypes = []string{KUBEADM, SKUBA}

type Bootstrapper interface {
	Init(cluster *data.Cluster) error
	Bootstrap(node *data.Node) error
	Join(node *data.Node) error
}

type BootstrapperInstaller interface {
	InstallRequirements() error
	CheckRequirements() error
}
