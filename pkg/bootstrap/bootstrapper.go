package bootstrap

import (
	"github.com/innobead/kubefire/pkg/data"
)

type Error error

const (
	KUBEADM = "kubeadm"
	SKUBA   = "skuba"
)

var BuiltinTypes = []string{KUBEADM, SKUBA}

type Bootstrapper interface {
	Init(cluster *data.Cluster) Error
	Bootstrap(node *data.Node) Error
	Join(node *data.Node) Error
}

type BootstrapperInstaller interface {
	Install() Error
	Check() Error
}
