package bootstrap

import (
	"github.com/innobead/kubefire/pkg/data"
	"github.com/innobead/kubefire/pkg/node"
	utilssh "github.com/innobead/kubefire/pkg/util/ssh"
	"github.com/sirupsen/logrus"
	"path"
)

const (
	KUBEADM = "kubeadm"
	SKUBA   = "skuba"
	K3S     = "k3s"
)

var BuiltinTypes = []string{
	KUBEADM,
	SKUBA,
	K3S,
}

type Bootstrapper interface {
	Deploy(cluster *data.Cluster, before func() error) error
	DownloadKubeConfig(cluster *data.Cluster, destDir string) (string, error)
	Prepare(force bool) error
}

func IsValid(bootstrapper string) bool {
	switch bootstrapper {
	case KUBEADM, SKUBA, K3S:
		return true
	default:
		return false
	}
}

func New(bootstrapper string, nodeManager node.Manager) Bootstrapper {
	switch bootstrapper {
	case SKUBA:
		return NewSkubaBootstrapper(nodeManager)

	case KUBEADM, "":
		return NewKubeadmBootstrapper(nodeManager)

	case K3S:
		return NewK3sBootstrapper(nodeManager)

	default:
		return nil
	}
}

func downloadKubeConfig(nodeManager node.Manager, cluster *data.Cluster, remoteKubeConfigPath string, destDir string) (string, error) {
	logrus.Infof("downloading the kubeconfig of cluster (%s)", cluster.Name)

	firstMaster, err := nodeManager.GetNode(node.Name(cluster.Name, node.Master, 1))
	if err != nil {
		return "", err
	}

	sshClient, err := utilssh.NewClient(
		firstMaster.Name,
		cluster.Spec.Prikey,
		"root",
		firstMaster.Status.IPAddresses,
		nil,
	)
	if err != nil {
		return "", err
	}
	defer sshClient.Close()

	destPath := cluster.Spec.LocalKubeConfig()

	if destDir != "" {
		destPath = path.Join(destDir, "admin.conf")
	}

	logrus.Infof("saved the kubeconfig of cluster (%s) to %s", cluster.Name, destPath)

	if remoteKubeConfigPath == "" {
		remoteKubeConfigPath = "/etc/kubernetes/admin.conf"
	}

	if err := sshClient.Download(remoteKubeConfigPath, destPath); err != nil {
		return "", err
	}

	return destPath, nil
}
