package bootstrap

import (
	"bytes"
	"fmt"
	"github.com/avast/retry-go"
	"github.com/hashicorp/go-multierror"
	"github.com/innobead/kubefire/internal/config"
	"github.com/innobead/kubefire/pkg/bootstrap/versionfinder"
	pkgconfig "github.com/innobead/kubefire/pkg/config"
	"github.com/innobead/kubefire/pkg/constants"
	"github.com/innobead/kubefire/pkg/data"
	"github.com/innobead/kubefire/pkg/node"
	"github.com/innobead/kubefire/pkg/script"
	utilssh "github.com/innobead/kubefire/pkg/util/ssh"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"strings"
	"sync"
	"time"
)

type KubeadmExtraOptions struct {
	InitOptions              []string `json:"init_options"`
	ApiServerOptions         []string `json:"api_server_options"`
	ControllerManagerOptions []string `json:"controller_manager_options"`
	SchedulerOptions         []string `json:"scheduler_options"`
}

type KubeadmBootstrapper struct {
	nodeManager   node.Manager
	versionFinder versionfinder.Finder
	configManager pkgconfig.Manager
}

func NewKubeadmBootstrapper() *KubeadmBootstrapper {
	return &KubeadmBootstrapper{}
}

func (k *KubeadmExtraOptions) generateKubeadmInitOptions() []string {
	var options []string
	for _, o := range k.InitOptions {
		if !strings.HasPrefix(o, "--") {
			o = "--" + o
		}

		options = append(options, o)
	}

	return options
}

func (k *KubeadmExtraOptions) generateControlPlaneComponentOptions(cpOptions *[]string) []string {
	var options []string
	for _, o := range *cpOptions {
		if strings.HasPrefix(o, "--") {
			o = strings.TrimPrefix(o, "--")
		}

		options = append(options, o)
	}

	return options
}

func (k *KubeadmBootstrapper) SetConfigManager(configManager pkgconfig.Manager) {
	k.configManager = configManager
}

func (k *KubeadmBootstrapper) SetVersionFinder(versionFinder versionfinder.Finder) {
	k.versionFinder = versionFinder
}

func (k *KubeadmBootstrapper) SetNodeManager(nodeManager node.Manager) {
	k.nodeManager = nodeManager
}

func (k *KubeadmBootstrapper) Deploy(cluster *data.Cluster, before func() error) error {
	if before != nil {
		if err := before(); err != nil {
			return err
		}
	}

	extraOptions := KubeadmExtraOptions{}
	if err := cluster.Spec.ParseExtraOptions(&extraOptions); err != nil {
		return err
	}

	if err := k.nodeManager.WaitNodesRunning(cluster.Name, 5); err != nil {
		return errors.WithMessage(err, "some nodes are not running")
	}

	if err := k.init(cluster); err != nil {
		return err
	}

	firstMaster, err := k.nodeManager.GetNode(node.Name(cluster.Name, node.Master, 1))
	if err != nil {
		return err
	}

	firstMaster.Spec.Cluster = &cluster.Spec

	joinCmd, err := k.bootstrap(firstMaster, len(cluster.Nodes) == 1, &extraOptions)
	if err != nil {
		return err
	}

	nodes, err := k.nodeManager.ListNodes(cluster.Name)
	if err != nil {
		return err
	}

	if len(nodes) == 0 {
		return errors.New("no nodes available")
	}

	for _, n := range nodes {
		if n.Name == firstMaster.Name {
			continue
		}
		n.Spec.Cluster = &cluster.Spec

		if err := k.join(n, joinCmd); err != nil {
			return err
		}
	}

	return nil
}

func (k *KubeadmBootstrapper) DownloadKubeConfig(cluster *data.Cluster, destDir string) (string, error) {
	return downloadKubeConfig(k.nodeManager, cluster, "", destDir)
}

func (k *KubeadmBootstrapper) Prepare(cluster *data.Cluster, force bool) error {
	return nil
}

func (k *KubeadmBootstrapper) Type() string {
	return constants.KUBEADM
}

func (k *KubeadmBootstrapper) init(cluster *data.Cluster) error {
	logrus.WithField("cluster", cluster.Name).Infoln("initializing cluster")

	bootstrapperVersion, err := getSupportedBootstrapperVersion(k.versionFinder, k.configManager, k, cluster.Spec.Version)
	if err != nil {
		return err
	}
	kubeadmBootstrapperVersion := bootstrapperVersion.(*pkgconfig.KubeadmBootstrapperVersion)

	wgInitNodes := sync.WaitGroup{}
	wgInitNodes.Add(len(cluster.Nodes))

	chErr := make(chan error, len(cluster.Nodes))

	for _, n := range cluster.Nodes {
		logrus.WithField("node", n.Name).Infoln("initializing node")

		go func(n *data.Node) {
			defer wgInitNodes.Done()

			_ = retry.Do(func() error {
				sshClient, err := utilssh.NewClient(
					n.Name,
					cluster.Spec.Prikey,
					"root",
					n.Status.IPAddresses,
					nil,
				)
				if err != nil {
					return err
				}
				defer sshClient.Close()

				cmds := []string{
					"swapoff -a",
					fmt.Sprintf("curl -sSLO %s", script.RemoteScriptUrl(script.InstallPrerequisitesKubeadm)),
					fmt.Sprintf("chmod +x %s", script.InstallPrerequisitesKubeadm),
					fmt.Sprintf(
						"%s ./%s",
						config.KubeadmVersionsEnvVars(
							kubeadmBootstrapperVersion.BootstrapperVersion,
							kubeadmBootstrapperVersion.KubeReleaseVersion,
							kubeadmBootstrapperVersion.CrictlVersion,
						).String(),
						script.InstallPrerequisitesKubeadm,
					),
					"sysctl -w net.ipv4.ip_forward=1",
					`echo "net.ipv4.ip_forward = 1" >> /etc/sysctl.conf`,
					`echo "0.0.0.0 $(hostname)" >> /etc/hosts`,
					`echo "export CONTAINER_RUNTIME_ENDPOINT=unix:///enabled/containerd/containerd.sock" >> /etc/profile.d/containerd.sh`,
					"kubeadm init phase preflight -v 5",
				}

				err = sshClient.Run(nil, nil, cmds...)
				if err != nil {
					chErr <- errors.WithMessagef(err, "failed on node (%s)", n.Name)
				}

				return nil
			},
				retry.Delay(10*time.Second),
				retry.MaxDelay(1*time.Minute),
			)
		}(n)
	}

	logrus.Info("waiting all nodes initialization finished")

	wgInitNodes.Wait()
	close(chErr)

	err = nil
	for {
		e, ok := <-chErr
		if !ok {
			break
		}

		if e != nil {
			err = multierror.Append(err, e)
		}
	}

	return err
}

func (k *KubeadmBootstrapper) bootstrap(node *data.Node, isSingleNode bool, options *KubeadmExtraOptions) (joinCmd string, err error) {
	logrus.WithField("node", node.Name).Infoln("bootstrapping the first master node")

	sshClient, err := utilssh.NewClient(
		node.Name,
		node.Spec.Cluster.Prikey,
		"root",
		node.Status.IPAddresses,
		nil,
	)
	if err != nil {
		return "", err
	}
	defer sshClient.Close()

	joinCmdBuf := bytes.Buffer{}
	ignoreErrors := []string{
		"FileAvailable--etc-kubernetes-manifests-kube-apiserver.yaml",
		"FileAvailable--etc-kubernetes-manifests-kube-controller-manager.yaml",
		"FileAvailable--etc-kubernetes-manifests-kube-scheduler.yaml",
	}

	cmds := []struct {
		cmdline string
		before  utilssh.Callback
		after   utilssh.Callback
	}{
		{
			cmdline: fmt.Sprintf(
				`kubeadm init phase control-plane all -v 5 --apiserver-extra-args="%s" --controller-manager-extra-args="%s" --scheduler-extra-args="%s"`,
				strings.Join(options.generateControlPlaneComponentOptions(&options.ApiServerOptions), ","),
				strings.Join(options.generateControlPlaneComponentOptions(&options.ControllerManagerOptions), ","),
				strings.Join(options.generateControlPlaneComponentOptions(&options.SchedulerOptions), ","),
			),
			before: func(session *ssh.Session) bool {
				logrus.Info("running kubeadm init")
				return true
			},
		},
		{
			cmdline: fmt.Sprintf(
				"kubeadm init -v 5 --skip-phases='control-plane' --ignore-preflight-errors='%s' %s",
				strings.Join(ignoreErrors, ","),
				strings.Join(options.generateKubeadmInitOptions(), " "),
			),
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
					logrus.WithField("node", node.Name).Infoln("untainting the master node")
				}

				return isSingleNode
			},
		},
	}

	for _, c := range cmds {
		err := sshClient.Run(c.before, c.after, c.cmdline)
		if err != nil {
			return "", errors.WithStack(err)
		}
	}

	return strings.TrimSuffix(joinCmdBuf.String(), "\n"), nil
}

func (k *KubeadmBootstrapper) join(node *data.Node, joinCmd string) error {
	logrus.WithField("node", node.Name).Infoln("joining node")

	sshClient, err := utilssh.NewClient(
		node.Name,
		node.Spec.Cluster.Prikey,
		"root",
		node.Status.IPAddresses,
		nil,
	)
	if err != nil {
		return err
	}
	defer sshClient.Close()

	logrus.Infof("running join command (%s)", joinCmd)

	if err := sshClient.Run(nil, nil, fmt.Sprintf("%s -v 5", joinCmd)); err != nil {
		return errors.WithStack(err)
	}

	return nil
}
