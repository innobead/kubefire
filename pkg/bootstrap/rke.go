package bootstrap

import (
	"context"
	"fmt"
	"github.com/goccy/go-yaml"
	"github.com/innobead/kubefire/internal/config"
	pkgconfig "github.com/innobead/kubefire/pkg/config"
	"github.com/innobead/kubefire/pkg/constants"
	"github.com/innobead/kubefire/pkg/data"
	"github.com/innobead/kubefire/pkg/node"
	"github.com/innobead/kubefire/pkg/script"
	"github.com/innobead/kubefire/pkg/util"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)
import "github.com/thoas/go-funk"

type RKEExtraOptions struct {
	ClusterConfigFile string `json:"cluster_config_file"`
	KubernetesVersion string `json:"kubernetes_version"`
}

type RKEBootstrapper struct {
	nodeManager node.Manager
}

func NewRKEBootstrapper() *RKEBootstrapper {
	return &RKEBootstrapper{}
}

func (k *RKEBootstrapper) SetNodeManager(nodeManager node.Manager) {
	k.nodeManager = nodeManager
}

func (k *RKEBootstrapper) Deploy(cluster *data.Cluster, before func() error) error {
	if before != nil {
		if err := before(); err != nil {
			return err
		}
	}

	extraOptions := RKEExtraOptions{}
	if err := cluster.Spec.ParseExtraOptions(&extraOptions); err != nil {
		return err
	}

	if err := k.nodeManager.WaitNodesRunning(cluster.Name, 5); err != nil {
		return errors.WithMessage(err, "some nodes are not running")
	}

	if err := k.init(cluster, &extraOptions); err != nil {
		return err
	}

	firstMaster, err := k.nodeManager.GetNode(node.Name(cluster.Name, node.Master, 1))
	if err != nil {
		return err
	}
	if _, err := k.bootstrap(firstMaster, &extraOptions); err != nil {
		return err
	}

	return nil
}

func (k *RKEBootstrapper) DownloadKubeConfig(cluster *data.Cluster, destDir string) (string, error) {
	downloadedKubeConfigPath := filepath.Join(cluster.Spec.LocalClusterDir(), "kube_config_cluster.rke.yaml")

	srcFile, err := os.Open(downloadedKubeConfigPath)
	if err != nil {
		return "", errors.WithStack(err)
	}

	destPath := cluster.Spec.LocalKubeConfig()
	if destDir != "" {
		destPath = path.Join(destDir, "admin.conf")
	}
	destFile, err := os.Create(destPath)
	if err != nil {
		return "", errors.WithStack(err)
	}

	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return "", errors.WithStack(err)
	}

	return destPath, nil
}

func (k *RKEBootstrapper) Prepare(cluster *data.Cluster, force bool) error {
	return k.installRKEExecutables(cluster.Spec.Version, force)
}

func (k *RKEBootstrapper) Type() string {
	return constants.RKE
}

func (k *RKEBootstrapper) init(cluster *data.Cluster, extraOptions *RKEExtraOptions) error {
	cmds := []string{
		"swapoff -a",
		fmt.Sprintf("curl -sfSLO %s", script.RemoteScriptUrl(script.InstallPrerequisitesRKE)),
		fmt.Sprintf("chmod +x %s", script.InstallPrerequisitesRKE),
		fmt.Sprintf("./%s node", script.InstallPrerequisitesRKE),
	}

	if err := initNodes(cluster, cmds); err != nil {
		return err
	}

	// generate cluster.yaml in cluster folder
	configPath := k.rkeClusterConfigPath(&cluster.Spec)

	logrus.WithField("cluster", cluster.Name).Infof("generating RKE cluster.yaml (%s)\n", configPath)

	type Node struct {
		Address    string   `json:"address"`
		User       string   `json:"user"`
		Role       []string `json:"role"`
		SshKeyPath string   `json:"ssh_key_path"`
		Port       int      `json:"port"`
	}

	var nodes []Node
	for _, n := range cluster.Nodes {
		var node = Node{
			Address:    n.Status.IPAddresses,
			User:       "root",
			SshKeyPath: cluster.Spec.Prikey,
			Port:       22,
		}

		if n.IsMaster() {
			node.Role = []string{"controlplane", "etcd"}

			if len(cluster.Nodes) == 1 {
				node.Role = append(node.Role, "worker")
			}
		} else {
			node.Role = []string{"worker"}
		}

		nodes = append(
			nodes,
			node,
		)
	}

	clusterConfig := map[string]interface{}{
		"nodes":        nodes,
		"cluster_name": cluster.Name,
	}
	if extraOptions.KubernetesVersion != "" {
		clusterConfig["kubernetes_version"] = extraOptions.KubernetesVersion
	}

	bytes, err := yaml.Marshal(clusterConfig)
	if err != nil {
		return errors.WithStack(err)
	}
	if err = ioutil.WriteFile(configPath, bytes, 0755); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (k *RKEBootstrapper) bootstrap(node *data.Node, extraOptions *RKEExtraOptions) (token string, err error) {
	cluster := node.Spec.Cluster
	configPath := k.rkeClusterConfigPath(cluster)

	if err := k.mergeClusterConfig(configPath, extraOptions); err != nil {
		return "", err
	}

	logrus.WithField("cluster", cluster.Name).Infof("deploying RKE cluster by using %s\n", configPath)

	cmdline := fmt.Sprintf("rke up --config %s", configPath)
	cmdArgs := strings.Split(cmdline, " ")

	cmd := util.UpdateCommandDefaultLogWithInfo(
		exec.CommandContext(
			context.Background(),
			cmdArgs[0],
			cmdArgs[1:]...,
		),
	)
	cmd.Dir = cluster.LocalClusterDir()

	if err := cmd.Run(); err != nil {
		return "", errors.WithStack(err)
	}

	return "", nil
}

func (k *RKEBootstrapper) mergeClusterConfig(configPath string, extraOptions *RKEExtraOptions) error {
	if extraOptions.ClusterConfigFile == "" {
		return nil
	}

	logrus.Infof("merging the cluster config (%s) with the user provided cluster config (%s)\n", configPath, extraOptions.ClusterConfigFile)

	// read the generated cluster config
	bytes, err := ioutil.ReadFile(configPath)
	if err != nil {
		return errors.WithStack(err)
	}
	clusterConfig := map[string]interface{}{}
	if err := yaml.Unmarshal(bytes, &clusterConfig); err != nil {
		return errors.WithStack(err)
	}

	// read the user provided cluster config
	bytes, err = ioutil.ReadFile(extraOptions.ClusterConfigFile)
	if err != nil {
		return errors.WithStack(err)
	}
	userClusterConfig := map[string]interface{}{}
	if err := yaml.Unmarshal(bytes, &userClusterConfig); err != nil {
		return errors.WithStack(err)
	}

	// merge
	ignoredKeys := []string{"nodes", "cluster_name", "kubernetes_version"}
	for k, v := range userClusterConfig {
		if funk.Contains(k, ignoredKeys) {
			continue
		}

		clusterConfig[k] = v
	}
	bytes, err = yaml.Marshal(&clusterConfig)
	if err != nil {
		return errors.WithStack(err)
	}
	err = ioutil.WriteFile(configPath, bytes, 0755)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (k *RKEBootstrapper) installRKEExecutables(version string, force bool) error {
	scripts := []script.Type{
		script.InstallPrerequisitesRKE,
	}

	for _, s := range scripts {
		if err := script.Download(s, config.TagVersion, force); err != nil {
			return err
		}

		if err := script.Run(s, config.TagVersion, func(cmd *exec.Cmd) error {
			cmd.Env = append(
				cmd.Env,
				config.RKEVersionsEnvVars(version)...,
			)

			return nil
		}); err != nil {
			return err
		}
	}

	return nil
}

func (k *RKEBootstrapper) rkeClusterConfigPath(cluster *pkgconfig.Cluster) string {
	return path.Join(cluster.LocalClusterDir(), "cluster.rke.yaml")
}
