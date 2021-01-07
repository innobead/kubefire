package bootstrap

import (
	"fmt"
	"github.com/goccy/go-yaml"
	"github.com/innobead/kubefire/internal/config"
	"github.com/innobead/kubefire/pkg/constants"
	"github.com/innobead/kubefire/pkg/data"
	"github.com/innobead/kubefire/pkg/node"
	"github.com/innobead/kubefire/pkg/script"
	"github.com/innobead/kubefire/pkg/util"
	utilssh "github.com/innobead/kubefire/pkg/util/ssh"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"reflect"
	"strings"
)

type RKE2ExtraOptions struct {
	ServerInstallOptions []string `json:"server_install_options"`
	AgentInstallOptions  []string `json:"agent_install_options"`
	ExtraOptions         []string `json:"extra_options"`
}

type RKE2Bootstrapper struct {
	nodeManager node.Manager
}

func NewRKE2Bootstrapper() *RKE2Bootstrapper {
	return &RKE2Bootstrapper{}
}

func (r *RKE2Bootstrapper) SetNodeManager(nodeManager node.Manager) {
	r.nodeManager = nodeManager
}

func (r *RKE2Bootstrapper) Deploy(cluster *data.Cluster, before func() error) error {
	if before != nil {
		if err := before(); err != nil {
			return err
		}
	}

	extraOptions := RKE2ExtraOptions{
		ExtraOptions: config.RKE2VersionsEnvVars(cluster.Spec.Version, ""),
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

func (r *RKE2Bootstrapper) DownloadKubeConfig(cluster *data.Cluster, destDir string) (string, error) {
	return downloadKubeConfig(r.nodeManager, cluster, "/etc/rancher/rke2/rke2.yaml", destDir)
}

func (r *RKE2Bootstrapper) Prepare(cluster *data.Cluster, force bool) error {
	return nil
}

func (r *RKE2Bootstrapper) Type() string {
	return constants.RKE2
}

func (r *RKE2Bootstrapper) init(cluster *data.Cluster) error {
	cmds := []string{
		"swapoff -a",
		fmt.Sprintf("curl -sfSLO %s", script.RemoteScriptUrl(script.InstallPrerequisitesRKE2)),
		fmt.Sprintf("chmod +x %s", script.InstallPrerequisitesRKE2),
		fmt.Sprintf("%s ./%s install_rke2", config.RKE2VersionsEnvVars(cluster.Spec.Version, "").String(), script.InstallPrerequisitesRKE2),
	}

	return initNodes(cluster, cmds)
}

func (r *RKE2Bootstrapper) bootstrap(node *data.Node, isSingleNode bool, extraOptions *RKE2ExtraOptions) (token string, err error) {
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

	deployConfigValue, err := createRKK2Config(deployCmdOpts)
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
				config.RKE2VersionsEnvVars(node.Spec.Cluster.Version, deployConfigValue).String(),
				script.InstallPrerequisitesRKE2,
			),
		},
		{
			cmdline: "rke2-install.sh",
		},
		{
			cmdline: "systemctl enable rke2-server.service",
		},
		{
			cmdline: "systemctl start rke2-server.service",
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

func (r *RKE2Bootstrapper) join(node *data.Node, apiServerAddress string, joinToken string, extraOptions *RKE2ExtraOptions) error {
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
	cmd := "INSTALL_RKE2_TYPE=server rke2-install.sh"
	systemdService := "rke2-server.service"
	if node.IsMaster() {
		if len(extraOptions.ServerInstallOptions) > 0 {
			deployCmdOpts = append(deployCmdOpts, extraOptions.ServerInstallOptions...)
		}
	} else {
		cmd = "INSTALL_RKE2_TYPE=agent rke2-install.sh"
		systemdService = "rke2-agent.service"
		if len(extraOptions.AgentInstallOptions) > 0 {
			deployCmdOpts = append(deployCmdOpts, extraOptions.AgentInstallOptions...)
		}
	}

	deployConfigValue, err := createRKK2Config(deployCmdOpts)
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
				config.RKE2VersionsEnvVars(node.Spec.Cluster.Version, deployConfigValue).String(),
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

func createRKK2Config(options []string) (string, error) {
	cfg := map[string]interface{}{}

	for _, str := range options {
		str = strings.TrimPrefix(str, "--")
		opt := strings.SplitN(str, "=", 2)

		if len(opt) != 2 {
			return "", errors.New(fmt.Sprintf("ignored the invalid option, %s", str))
		}

		if _, exist := cfg[opt[0]]; exist {
			switch reflect.TypeOf(cfg[opt[0]]).Kind() {
			case reflect.Slice, reflect.Array:
				cfg[opt[0]] = append(cfg[opt[0]].([]interface{}), opt[1])
			default:
				cfg[opt[0]] = []interface{}{opt[1]}
			}
		}

		cfg[opt[0]] = opt[1]
	}

	if len(cfg) == 0 {
		return "", nil
	}

	bytes, err := yaml.Marshal(cfg)
	if err != nil {
		return "", errors.WithStack(err)
	}

	return string(bytes), nil
}
