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
	Deploy(cluster *data.Cluster) error
}
