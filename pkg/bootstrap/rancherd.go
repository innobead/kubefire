package bootstrap

import (
	"fmt"
	"github.com/innobead/kubefire/internal/config"
	"github.com/innobead/kubefire/pkg/constants"
	"github.com/innobead/kubefire/pkg/data"
	"github.com/innobead/kubefire/pkg/node"
	"github.com/innobead/kubefire/pkg/script"
	"github.com/innobead/kubefire/pkg/util"
	utilssh "github.com/innobead/kubefire/pkg/util/ssh"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"strings"
)

type RancherdExtraOptions struct {
	ServerInstallOptions []string `json:"server_install_options"`
	AgentInstallOptions  []string `json:"agent_install_options"`
	ExtraOptions         []string `json:"extra_options"`
}

type RancherdBootstrapper struct {
	*RKE2Bootstrapper
}

func NewRancherdBootstrapper() *RancherdBootstrapper {
	return &RancherdBootstrapper{
		RKE2Bootstrapper: NewRKE2Bootstrapper(),
	}
}

func (r *RancherdBootstrapper) SetNodeManager(nodeManager node.Manager) {
	r.nodeManager = nodeManager
}

func (r *RancherdBootstrapper) Deploy(cluster *data.Cluster, before func() error) error {
	if before != nil {
		if err := before(); err != nil {
			return err
		}
	}

	extraOptions := RancherdExtraOptions{
		ExtraOptions: config.RancherdVersionsEnvVars(cluster.Spec.Version, ""),
	}
	if err := cluster.Spec.ParseExtraOptions(&extraOptions); err != nil {
		return err
	}

	if err := r.nodeManager.WaitNodesRunning(cluster.Name, 5); err != nil {
		return errors.WithMessage(err, "some nodes are not running")
	}

	if err := r.init(cluster); err != nil {
		return err
	}

	firstMaster, err := r.nodeManager.GetNode(node.Name(cluster.Name, node.Master, 1))
	if err != nil {
		return err
	}

	firstMaster.Spec.Cluster = &cluster.Spec

	joinToken, err := r.bootstrap(firstMaster, len(cluster.Nodes) == 1, &extraOptions)
	if err != nil {
		return err
	}

	nodes, err := r.nodeManager.ListNodes(cluster.Name)
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

		if err := r.join(n, firstMaster.Status.IPAddresses, joinToken, &extraOptions); err != nil {
			return err
		}
	}

	return nil
}

func (r *RancherdBootstrapper) DownloadKubeConfig(cluster *data.Cluster, destDir string) (string, error) {
	return r.RKE2Bootstrapper.DownloadKubeConfig(cluster, destDir)
}

func (r *RancherdBootstrapper) Prepare(cluster *data.Cluster, force bool) error {
	return r.RKE2Bootstrapper.Prepare(cluster, force)
}

func (r *RancherdBootstrapper) Type() string {
	return constants.RANCHERD
}

func (r *RancherdBootstrapper) init(cluster *data.Cluster) error {
	cmds := []string{
		"swapoff -a",
		fmt.Sprintf("curl -sfSLO %s", script.RemoteScriptUrl(script.InstallPrerequisitesRKE2)),
		fmt.Sprintf("chmod +x %s", script.InstallPrerequisitesRKE2),
		fmt.Sprintf("%s ./%s install_rancherd", config.RancherdVersionsEnvVars(cluster.Spec.Version, "").String(), script.InstallPrerequisitesRKE2),
	}

	return initNodes(cluster, cmds)
}

func (r *RancherdBootstrapper) bootstrap(node *data.Node, isSingleNode bool, extraOptions *RancherdExtraOptions) (token string, err error) {
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

	joinToken := util.GenerateRandomStr(8)
	deployCmdOpts := []string{
		fmt.Sprintf("--bind-address=%s", node.Status.IPAddresses),
		fmt.Sprintf("--token=%s", joinToken),
	}

	if extraOptions.ServerInstallOptions != nil {
		deployCmdOpts = append(deployCmdOpts, extraOptions.ServerInstallOptions...)
	}

	deployConfigValue, err := createRKE2Config(deployCmdOpts)
	if err != nil {
		return "", err
	}

	cmds := []struct {
		cmdline string
		before  utilssh.Callback
	}{
		{
			cmdline: fmt.Sprintf(
				"%s ./%s create_config",
				config.RancherdVersionsEnvVars(node.Spec.Cluster.Version, deployConfigValue).String(),
				script.InstallPrerequisitesRKE2,
			),
		},
		{
			cmdline: "rancherd-install.sh",
		},
		{
			cmdline: "systemctl enable rancherd-server.service",
		},
		{
			cmdline: "systemctl start rancherd-server.service",
		},
	}

	for _, c := range cmds {
		err := sshClient.Run(c.before, nil, c.cmdline)
		if err != nil {
			return "", errors.WithStack(err)
		}
	}

	return strings.TrimSuffix(joinToken, "\n"), nil
}

func (r *RancherdBootstrapper) join(node *data.Node, apiServerAddress string, joinToken string, extraOptions *RancherdExtraOptions) error {
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

	deployCmdOpts := []string{
		fmt.Sprintf("--server=https://%s:9345", apiServerAddress),
		fmt.Sprintf("--token=%s", joinToken),
	}
	cmd := "INSTALL_RKE2_TYPE=server rancherd-install.sh"
	systemdService := "rancherd-server.service"
	if node.IsMaster() {
		if len(extraOptions.ServerInstallOptions) > 0 {
			deployCmdOpts = append(deployCmdOpts, extraOptions.ServerInstallOptions...)
		}
	} else {
		cmd = "INSTALL_RKE2_TYPE=agent rancherd-install.sh"
		systemdService = "rancherd-agent.service"
		if len(extraOptions.AgentInstallOptions) > 0 {
			deployCmdOpts = append(deployCmdOpts, extraOptions.AgentInstallOptions...)
		}
	}

	deployConfigValue, err := createRKE2Config(deployCmdOpts)
	if err != nil {
		return err
	}

	cmds := []struct {
		cmdline string
		before  utilssh.Callback
	}{
		{
			cmdline: fmt.Sprintf(
				"%s ./%s create_config",
				config.RancherdVersionsEnvVars(node.Spec.Cluster.Version, deployConfigValue).String(),
				script.InstallPrerequisitesRKE2,
			),
		},
		{
			cmdline: cmd,
		},
		{
			cmdline: fmt.Sprintf("systemctl enable %s", systemdService),
		},
		{
			cmdline: fmt.Sprintf("systemctl start %s", systemdService),
		},
	}

	for _, c := range cmds {
		err := sshClient.Run(c.before, nil, c.cmdline)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}
