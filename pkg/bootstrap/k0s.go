package bootstrap

import (
	"bytes"
	"fmt"
	"github.com/innobead/kubefire/internal/config"
	pkgconfig "github.com/innobead/kubefire/pkg/config"
	"github.com/innobead/kubefire/pkg/constants"
	"github.com/innobead/kubefire/pkg/data"
	"github.com/innobead/kubefire/pkg/node"
	"github.com/innobead/kubefire/pkg/script"
	utilssh "github.com/innobead/kubefire/pkg/util/ssh"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"text/template"
)

const configTemplate = `apiVersion: k0s.k0sproject.io/v1beta1
kind: Cluster
metadata:
  name: k0s
spec:
  api:
    externalAddress: {{.BindAddress}}
    address: {{.BindAddress}}
    sans:
    - {{.BindAddress}}
`

type K0sExtraOptions struct {
	ClusterConfigFile    string   `json:"cluster_config_file"`
	ServerInstallOptions []string `json:"server_install_options"`
	WorkerInstallOptions []string `json:"worker_install_options"`
	ExtraOptions         []string `json:"extra_options"`
}

type K0sBootstrapper struct {
	nodeManager node.Manager
}

func NewK0sBootstrapper() *K0sBootstrapper {
	return &K0sBootstrapper{}
}

func (k *K0sBootstrapper) SetNodeManager(nodeManager node.Manager) {
	k.nodeManager = nodeManager
}

func (k *K0sBootstrapper) Deploy(cluster *data.Cluster, before func() error) error {
	if before != nil {
		if err := before(); err != nil {
			return err
		}
	}

	extraOptions := K0sExtraOptions{
		ExtraOptions: config.K0sVersionsEnvVars(cluster.Spec.Version, "", ""),
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

	serverJoinToken, workerJoinToken, err := k.bootstrap(firstMaster, len(cluster.Nodes) == 1, &extraOptions)
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

		if err := k.join(n, serverJoinToken, workerJoinToken, &extraOptions); err != nil {
			return err
		}
	}

	return nil
}

func (k *K0sBootstrapper) DownloadKubeConfig(cluster *data.Cluster, destDir string) (string, error) {
	return downloadKubeConfig(k.nodeManager, cluster, "/var/lib/k0s/pki/admin.conf", destDir)
}

func (k *K0sBootstrapper) Prepare(cluster *data.Cluster, force bool) error {
	return nil
}

func (k *K0sBootstrapper) Type() string {
	return constants.K0s
}

func (k *K0sBootstrapper) init(cluster *data.Cluster) error {
	cmds := []string{
		"swapoff -a",
		fmt.Sprintf("curl -sfSLO %s", script.RemoteScriptUrl(script.InstallPrerequisitesK0s)),
		fmt.Sprintf("chmod +x %s", script.InstallPrerequisitesK0s),
		fmt.Sprintf("%s ./%s install_k0s", config.K0sVersionsEnvVars(cluster.Spec.Version, "", "").String(), script.InstallPrerequisitesK0s),
	}

	return initNodes(cluster, cmds)
}

func (k *K0sBootstrapper) bootstrap(node *data.Node, isSingleNode bool, extraOptions *K0sExtraOptions) (serverToken string, workerToken string, err error) {
	logrus.WithField("node", node.Name).Infoln("bootstrapping the first master node")

	sshClient, err := utilssh.NewClient(
		node.Name,
		node.Spec.Cluster.Prikey,
		"root",
		node.Status.IPAddresses,
		nil,
	)
	if err != nil {
		return "", "", err
	}
	defer sshClient.Close()

	var deployCmdOpts []string
	if extraOptions.ServerInstallOptions != nil {
		deployCmdOpts = append(deployCmdOpts, extraOptions.ServerInstallOptions...)
	}
	//FIXME need to enable worker, otherwise the workloads will be pending due to no node available. Probably upstream issue.
	//if isSingleNode {
	deployCmdOpts = append(deployCmdOpts, "--enable-worker --no-taints")
	//}

	// create the default cluster config
	configPath := k.clusterConfigPath(node.Spec.Cluster)

	tmp, err := template.New("").Parse(configTemplate)
	if err != nil {
		return "", "", errors.WithStack(err)
	}

	file, err := os.Create(configPath)
	if err != nil {
		return "", "", errors.WithStack(err)
	}

	err = tmp.Execute(file, struct {
		BindAddress string
	}{
		BindAddress: node.Status.IPAddresses,
	})
	if err != nil {
		return "", "", errors.WithStack(err)
	}

	// merge the default cluster config with the user provided config
	if err := mergeClusterConfig(configPath, extraOptions.ClusterConfigFile, nil); err != nil {
		return "", "", err
	}

	rawBytes, err := ioutil.ReadFile(configPath)
	if err != nil {
		return "", "", errors.WithStack(err)
	}
	deployConfigValue := string(rawBytes)

	serverTokenBuf := bytes.Buffer{}
	workerTokenBuf := bytes.Buffer{}

	cmds := []struct {
		cmdline string
		before  utilssh.Callback
	}{
		{
			cmdline: fmt.Sprintf(
				"%s ./%s create_controller",
				config.K0sVersionsEnvVars(
					node.Spec.Cluster.Version,
					deployConfigValue,
					fmt.Sprintf("server -c %s %s", "/etc/k0s/config.yaml", strings.Join(deployCmdOpts, " ")),
				).String(),
				script.InstallPrerequisitesK0s,
			),
		},
		{
			// make sure the related CA generated before creating the token in the following commands
			cmdline: "sleep 30s",
		},
		{
			cmdline: "k0s token create --role=controller",
			before: func(session *ssh.Session) bool {
				session.Stdout = &serverTokenBuf
				return true
			},
		},
		{
			cmdline: "k0s token create --role=worker",
			before: func(session *ssh.Session) bool {
				session.Stdout = &workerTokenBuf
				return true
			},
		},
	}

	for _, c := range cmds {
		err := sshClient.Run(c.before, nil, c.cmdline)
		if err != nil {
			return "", "", errors.WithStack(err)
		}
	}

	return strings.TrimSuffix(serverTokenBuf.String(), "\n"), strings.TrimSuffix(workerTokenBuf.String(), "\n"), nil
}

func (k *K0sBootstrapper) join(node *data.Node, serverJoinToken string, workerJoinToken string, extraOptions *K0sExtraOptions) error {
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

	cmd := fmt.Sprintf(`server --enable-worker --no-taints %s`, serverJoinToken)
	if node.IsMaster() {
		if len(extraOptions.ServerInstallOptions) > 0 {
			deployCmdOpts = append(deployCmdOpts, extraOptions.ServerInstallOptions...)
		}
	} else {
		cmd = fmt.Sprintf("worker %s", workerJoinToken)
		if len(extraOptions.WorkerInstallOptions) > 0 {
			deployCmdOpts = append(deployCmdOpts, extraOptions.WorkerInstallOptions...)
		}
	}

	cmds := []struct {
		cmdline string
		before  utilssh.Callback
	}{
		{
			cmdline: fmt.Sprintf(
				"%s ./%s join_node",
				config.K0sVersionsEnvVars(
					node.Spec.Cluster.Version,
					"",
					fmt.Sprintf("%s %s", cmd, strings.Join(deployCmdOpts, " ")),
				).String(),
				script.InstallPrerequisitesK0s,
			),
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

func (k *K0sBootstrapper) clusterConfigPath(cluster *pkgconfig.Cluster) string {
	return path.Join(cluster.LocalClusterDir(), "cluster.k0s.yaml")
}
