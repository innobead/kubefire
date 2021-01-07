package bootstrap

import (
	"bytes"
	"fmt"
	"github.com/innobead/kubefire/internal/config"
	"github.com/innobead/kubefire/pkg/constants"
	"github.com/innobead/kubefire/pkg/data"
	"github.com/innobead/kubefire/pkg/node"
	"github.com/innobead/kubefire/pkg/script"
	utilssh "github.com/innobead/kubefire/pkg/util/ssh"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"strings"
)

type K3sExtraOptions struct {
	ServerInstallOptions []string `json:"server_install_options"`
	AgentInstallOptions  []string `json:"agent_install_options"`
	ExtraOptions         []string `json:"extra_options"`
}

type K3sBootstrapper struct {
	nodeManager node.Manager
}

func NewK3sBootstrapper() *K3sBootstrapper {
	return &K3sBootstrapper{}
}

func (k *K3sBootstrapper) SetNodeManager(nodeManager node.Manager) {
	k.nodeManager = nodeManager
}

func (k *K3sBootstrapper) Deploy(cluster *data.Cluster, before func() error) error {
	if before != nil {
		if err := before(); err != nil {
			return err
		}
	}

	extraOptions := K3sExtraOptions{
		ExtraOptions: config.K3sVersionsEnvVars(cluster.Spec.Version),
	}
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

	joinToken, err := k.bootstrap(firstMaster, len(cluster.Nodes) == 1, &extraOptions)
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

		if err := k.join(n, firstMaster.Status.IPAddresses, joinToken, &extraOptions); err != nil {
			return err
		}
	}

	return nil
}

func (k *K3sBootstrapper) DownloadKubeConfig(cluster *data.Cluster, destDir string) (string, error) {
	return downloadKubeConfig(k.nodeManager, cluster, "/etc/rancher/k3s/k3s.yaml", destDir)
}

func (k *K3sBootstrapper) Prepare(cluster *data.Cluster, force bool) error {
	return nil
}

func (k *K3sBootstrapper) Type() string {
	return constants.K3S
}

func (k *K3sBootstrapper) init(cluster *data.Cluster) error {
	cmds := []string{
		"swapoff -a",
		fmt.Sprintf("curl -sfSLO %s", script.RemoteScriptUrl(script.InstallPrerequisitesK3s)),
		fmt.Sprintf("chmod +x %s", script.InstallPrerequisitesK3s),
		fmt.Sprintf("%s ./%s", config.K3sVersionsEnvVars(cluster.Spec.Version).String(), script.InstallPrerequisitesK3s),
	}

	return initNodes(cluster, cmds)
}

func (k *K3sBootstrapper) bootstrap(node *data.Node, isSingleNode bool, extraOptions *K3sExtraOptions) (token string, err error) {
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

	deployCmdOpts := []string{
		fmt.Sprintf("--bind-address=%s", node.Status.IPAddresses),
	}
	tokenBuf := bytes.Buffer{}

	if !isSingleNode {
		deployCmdOpts = append(deployCmdOpts, "--cluster-init")
	}

	if extraOptions.ServerInstallOptions != nil {
		deployCmdOpts = append(deployCmdOpts, extraOptions.ServerInstallOptions...)
	}

	cmds := []struct {
		cmdline string
		before  utilssh.Callback
	}{
		{
			cmdline: fmt.Sprintf(`INSTALL_K3S_EXEC="%s" %s k3s-install.sh `, strings.Join(deployCmdOpts, " "), strings.Join(extraOptions.ExtraOptions, " ")),
		},
		{
			cmdline: "cat /var/lib/rancher/k3s/server/node-token",
			before: func(session *ssh.Session) bool {
				session.Stdout = &tokenBuf
				return true
			},
		},
	}

	for _, c := range cmds {
		err := sshClient.Run(c.before, nil, c.cmdline)
		if err != nil {
			return "", errors.WithStack(err)
		}
	}

	return strings.TrimSuffix(tokenBuf.String(), "\n"), nil
}

func (k *K3sBootstrapper) join(node *data.Node, apiServerAddress string, joinToken string, extraOptions *K3sExtraOptions) error {
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

	var deployCmdOpts []string
	cmd := fmt.Sprintf("K3S_URL=https://%s:6443 K3S_TOKEN=%s k3s-install.sh", apiServerAddress, joinToken)

	if node.IsMaster() {
		deployCmdOpts = append(deployCmdOpts, "--server")

		if len(extraOptions.ServerInstallOptions) > 0 {
			deployCmdOpts = append(deployCmdOpts, extraOptions.ServerInstallOptions...)
		}
	} else {
		if len(extraOptions.AgentInstallOptions) > 0 {
			deployCmdOpts = append(deployCmdOpts, extraOptions.AgentInstallOptions...)
		}
	}

	cmd = fmt.Sprintf(`INSTALL_K3S_EXEC="%s" %s %s`, strings.Join(deployCmdOpts, " "), strings.Join(extraOptions.ExtraOptions, " "), cmd)

	if err := sshClient.Run(nil, nil, cmd); err != nil {
		return errors.WithStack(err)
	}

	return nil
}
