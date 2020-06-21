package bootstrap

import (
	"bytes"
	"fmt"
	"github.com/hashicorp/go-multierror"
	"github.com/innobead/kubefire/pkg/data"
	"github.com/innobead/kubefire/pkg/node"
	"github.com/innobead/kubefire/pkg/script"
	"github.com/innobead/kubefire/pkg/util"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"sync"
)

type KubeadmBootstrapper struct {
	nodeManager node.Manager
}

func NewKubeadmBootstrapper(nodeManager node.Manager) *KubeadmBootstrapper {
	return &KubeadmBootstrapper{nodeManager: nodeManager}
}

func (k *KubeadmBootstrapper) Deploy(cluster *data.Cluster) error {
	if err := k.nodeManager.WaitNodesRunning(cluster.Name, 5); err != nil {
		return errors.WithMessage(err, "some nodes are not running")
	}

	sshConfig, err := util.CreateSSHClientConfig(cluster.Spec.Prikey, "root", nil)
	if err != nil {
		return err
	}

	if err := k.init(sshConfig, cluster); err != nil {
		return err
	}

	firstMaster, err := k.nodeManager.GetNode(node.NodeName(cluster.Name, node.Master, 1))
	if err != nil {
		return err
	}

	joinCmd, err := k.bootstrap(sshConfig, firstMaster, len(cluster.Nodes) == 0)
	if err != nil {
		return err
	}

	nodes, err := k.nodeManager.ListNodes(cluster.Name)
	if err != nil {
		return err
	}

	for _, n := range nodes {
		if err := k.join(sshConfig, n, joinCmd); err != nil {
			return err
		}
	}

	return nil
}

func (k *KubeadmBootstrapper) init(sshConfig *ssh.ClientConfig, cluster *data.Cluster) error {
	logrus.Infof("initializing cluster (%s)", cluster.Name)

	wgInitNodes := sync.WaitGroup{}
	wgInitNodes.Add(len(cluster.Nodes))

	chInitNodesErrors := make(chan error, len(cluster.Nodes))

	for _, n := range cluster.Nodes {
		logrus.Infof("initializing node (%s)", n.Name)

		go func(n *data.Node) {
			defer wgInitNodes.Done()

			session, err := util.CreateSSHSession(n.Status.IPAddresses, sshConfig)
			if err != nil {
				chInitNodesErrors <- err
				return
			}
			defer session.Close()

			cmdLines := []string{
				fmt.Sprintf("curl -sSLO %s", script.RemoteScriptUrl(script.InstallPrerequisitesKubeadm)),
				fmt.Sprintf("chmod +x %s", script.InstallPrerequisitesKubeadm),
				fmt.Sprintf("./%s", script.InstallPrerequisitesKubeadm),
				"sysctl -w net.ipv4.ip_forward=1",
				`echo "net.ipv4.ip_forward = 1" >> /etc/sysctl.conf`,
				"echo $(hostname) >> /etc/hosts",
				`echo "export CONTAINER_RUNTIME_ENDPOINT=unix:///run/containerd/containerd.sock"" >> /etc/profile.d/containerd.sh`,
				"kubeadm init phase preflight -v 5",
			}

			for _, c := range cmdLines {
				if err := session.Run(c); err != nil {
					chInitNodesErrors <- err
					return
				}
			}
		}(n)
	}

	logrus.Info("waiting all nodes initialization finished")
	wgInitNodes.Wait()
	close(chInitNodesErrors)

	var err error
	for {
		e, ok := <-chInitNodesErrors
		if !ok {
			break
		}

		err = multierror.Append(err, e)
	}

	return err
}

func (k *KubeadmBootstrapper) bootstrap(sshClientConfig *ssh.ClientConfig, node *data.Node, isSingleNode bool) (joinCmd string, err error) {
	logrus.Infof("bootstrapping the first master node (%s)", node.Name)

	session, err := util.CreateSSHSession(node.Status.IPAddresses, sshClientConfig)
	if err != nil {
		return "", err
	}
	defer session.Close()

	logrus.Info("running kubeadm init")

	if err := session.Run("kubeadm init -v 5"); err != nil {
		return "", errors.WithStack(err)
	}

	logrus.Info("creating the join command")

	joinCmdBuf := bytes.Buffer{}
	oldStdout := session.Stdout
	session.Stdout = &joinCmdBuf
	if err := session.Run("kubeadm token create --print-join-command"); err != nil {
		return "", errors.WithStack(err)
	}
	session.Stdout = oldStdout

	logrus.Info("applying CNI network")

	// kubectl apply -f https://docs.projectcalico.org/v3.14/manifests/calico.yaml
	if err := session.Run("kubectl create -f https://raw.githubusercontent.com/cilium/cilium/v1.8/install/kubernetes/quick-install.yaml"); err != nil {
		return "", errors.WithStack(err)
	}

	if isSingleNode {
		logrus.Infof("untainting the master node (%s)", node.Name)

		if err := session.Run("kubectl taint nodes --all node-role.kubernetes.io/master-"); err != nil {
			return "", errors.WithStack(err)
		}
	}

	return joinCmdBuf.String(), nil
}

func (k *KubeadmBootstrapper) join(sshClientConfig *ssh.ClientConfig, node *data.Node, joinCmd string) error {
	logrus.Infof("joining node (%s)", node.Name)

	session, err := util.CreateSSHSession(node.Status.IPAddresses, sshClientConfig)
	if err != nil {
		return err
	}
	defer session.Close()

	logrus.Infof("running join commnad (%s)", joinCmd)

	if err := session.Run(fmt.Sprintf("%s -v 5", joinCmd)); err != nil {
		return errors.WithStack(err)
	}

	return nil
}
