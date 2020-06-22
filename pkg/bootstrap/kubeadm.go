package bootstrap

import (
	"bytes"
	"fmt"
	"github.com/avast/retry-go"
	"github.com/hashicorp/go-multierror"
	"github.com/innobead/kubefire/pkg/data"
	"github.com/innobead/kubefire/pkg/node"
	"github.com/innobead/kubefire/pkg/script"
	"github.com/innobead/kubefire/pkg/util"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"sync"
	"time"
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
		if n.Name == firstMaster.Name {
			continue
		}

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

			_ = retry.Do(func() error {
				sshClient, err := util.CreateSSHClient(n.Status.IPAddresses, sshConfig)
				if err != nil {
					return err
				}
				defer sshClient.Close()

				cmdLines := []string{
					fmt.Sprintf("curl -sSLO %s", script.RemoteScriptUrl(script.InstallPrerequisitesKubeadm)),
					fmt.Sprintf("chmod +x %s", script.InstallPrerequisitesKubeadm),
					fmt.Sprintf("./%s", script.InstallPrerequisitesKubeadm),
					"sysctl -w net.ipv4.ip_forward=1",
					`echo "net.ipv4.ip_forward = 1" >> /etc/sysctl.conf`,
					`echo "0.0.0.0 $(hostname)" >> /etc/hosts`,
					`echo "export CONTAINER_RUNTIME_ENDPOINT=unix:///run/containerd/containerd.sock" >> /etc/profile.d/containerd.sh`,
					"kubeadm init phase preflight -v 5",
				}

				for _, c := range cmdLines {
					var err error

					session, err := util.CreateSSHSession(sshClient)
					if err != nil {
						return err
					}

					func() {
						defer func() {
							session.Close()
						}()

						if e := session.Run(c); e != nil {
							err = e
						}
					}()

					if err != nil {
						chInitNodesErrors <- errors.WithMessagef(err, "failed on node (%s)", n.Name)
					}
				}

				return nil
			}, retry.Delay(10*time.Second), retry.MaxDelay(1*time.Minute))
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

		if e != nil {
			err = multierror.Append(err, e)
		}
	}

	return err
}

func (k *KubeadmBootstrapper) bootstrap(sshClientConfig *ssh.ClientConfig, node *data.Node, isSingleNode bool) (joinCmd string, err error) {
	logrus.Infof("bootstrapping the first master node (%s)", node.Name)

	sshClient, err := util.CreateSSHClient(node.Status.IPAddresses, sshClientConfig)
	if err != nil {
		return "", err
	}
	defer sshClient.Close()

	joinCmdBuf := bytes.Buffer{}

	cmds := []struct {
		cmdline string
		before  func(session *ssh.Session) bool
		after   func(session *ssh.Session)
	}{
		{
			cmdline: "kubeadm init -v 5",
			before: func(session *ssh.Session) bool {
				logrus.Info("running kubeadm init")
				return true
			},
		},
		{
			cmdline: "kubeadm token create --print-join-command",
			before: func(session *ssh.Session) bool {
				logrus.Info("creating the join command")
				session.Stdout = &joinCmdBuf
				return true
			},
		},
		{
			cmdline: "KUBECONFIG=/etc/kubernetes/admin.conf kubectl create -f https://raw.githubusercontent.com/cilium/cilium/v1.8/install/kubernetes/quick-install.yaml",
			before: func(session *ssh.Session) bool {
				logrus.Info("applying CNI network")
				return true
			},
		},
		{
			cmdline: "KUBECONFIG=/etc/kubernetes/admin.conf kubectl taint nodes --all node-role.kubernetes.io/master-",
			before: func(session *ssh.Session) bool {
				if isSingleNode {
					logrus.Infof("untainting the master node (%s)", node.Name)
				}

				return isSingleNode
			},
		},
	}

	for _, cmd := range cmds {
		session, err := util.CreateSSHSession(sshClient)
		if err != nil {
			return "", err
		}

		if err := func() error {
			defer session.Close()

			if cmd.before != nil && cmd.before(session) {
				if err := session.Run(cmd.cmdline); err != nil {
					return err
				}

				if cmd.after != nil {
					cmd.after(session)
				}
			}

			return nil
		}(); err != nil {
			return "", errors.WithStack(err)
		}
	}

	return joinCmdBuf.String(), nil
}

func (k *KubeadmBootstrapper) join(sshClientConfig *ssh.ClientConfig, node *data.Node, joinCmd string) error {
	logrus.Infof("joining node (%s)", node.Name)

	sshClient, err := util.CreateSSHClient(node.Status.IPAddresses, sshClientConfig)
	if err != nil {
		return err
	}
	defer sshClient.Close()

	session, err := util.CreateSSHSession(sshClient)
	if err != nil {
		return err
	}
	defer session.Close()

	logrus.Infof("running join command (%s)", joinCmd)

	if err := session.Run(fmt.Sprintf("%s -v 5", joinCmd)); err != nil {
		return errors.WithStack(err)
	}

	return nil
}
