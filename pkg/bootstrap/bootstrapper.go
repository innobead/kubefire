package bootstrap

import (
	"github.com/innobead/kubefire/pkg/config"
	"github.com/innobead/kubefire/pkg/data"
	"github.com/innobead/kubefire/pkg/node"
	utilssh "github.com/innobead/kubefire/pkg/util/ssh"
	"github.com/sirupsen/logrus"
	"path/filepath"
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
	DownloadKubeConfig(cluster *data.Cluster, destDir string) error
}

func downloadKubeConfig(nodeManager node.Manager, cluster *data.Cluster, remoteKubeConfigPath string, destDir string) error {
	logrus.Infof("downloading the kubeconfig of cluster (%s)", cluster.Name)

	firstMaster, err := nodeManager.GetNode(node.NodeName(cluster.Name, node.Master, 1))
	if err != nil {
		return err
	}

	sshClient, err := utilssh.NewClient(
		firstMaster.Name,
		cluster.Spec.Prikey,
		"root",
		firstMaster.Status.IPAddresses,
		nil,
	)
	if err != nil {
		return err
	}
	defer sshClient.Close()

	if destDir == "" {
		destDir = config.LocalClusterDir(cluster.Name)
	}

	destPath := filepath.Join(destDir, "admin.conf")
	logrus.Infof("saved the kubeconfig of cluster (%s) to %s", cluster.Name, destPath)

	if remoteKubeConfigPath == "" {
		remoteKubeConfigPath = "/etc/kubernetes/admin.conf"
	}

	if err := sshClient.Download(remoteKubeConfigPath, destPath); err != nil {
		return err
	}

	return nil
}
