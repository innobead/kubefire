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

	init(cluster *data.Cluster) error
	bootstrap(node *data.Node) error
	join(node *data.Node) error
}

//
//$ lscpu | grep Virtualization
//Virtualization:      VT-x
//
//$ lsmod | grep kvm
//kvm_intel             200704  0
//kvm                   593920  1 kvm_intel
