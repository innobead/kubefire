package bootstrap

import (
	"context"
	"fmt"
	"github.com/avast/retry-go"
	"github.com/hashicorp/go-multierror"
	"github.com/innobead/kubefire/internal/config"
	"github.com/innobead/kubefire/pkg/constants"
	"github.com/innobead/kubefire/pkg/data"
	"github.com/innobead/kubefire/pkg/node"
	"github.com/innobead/kubefire/pkg/script"
	"github.com/innobead/kubefire/pkg/util"
	utilssh "github.com/innobead/kubefire/pkg/util/ssh"
	"github.com/pkg/errors"
	"os/exec"
	"path"
	"strings"
	"sync"
	"time"
)

type SkubaExtraOptions struct {
	RegisterCode string
}

type SkubaBootstrapper struct {
	nodeManager node.Manager
}

func NewSkubaBootstrapper(nodeManager node.Manager) *SkubaBootstrapper {
	return &SkubaBootstrapper{nodeManager: nodeManager}
}

func (s *SkubaBootstrapper) Deploy(cluster *data.Cluster, before func() error) error {
	if before != nil {
		if err := before(); err != nil {
			return err
		}
	}

	extraOptions := cluster.Spec.ParseExtraOptions(&SkubaExtraOptions{}).(SkubaExtraOptions)

	if err := s.nodeManager.WaitNodesRunning(cluster.Name, 5); err != nil {
		return errors.WithMessage(err, "some nodes are not running")
	}

	firstMaster, err := s.nodeManager.GetNode(node.Name(cluster.Name, node.Master, 1))
	if err != nil {
		return err
	}
	firstMaster.Spec.Cluster = &cluster.Spec

	clusterDir, err := s.init(cluster, firstMaster, &extraOptions)
	if err != nil {
		return err
	}

	err = s.bootstrap(firstMaster, clusterDir, len(cluster.Nodes) == 1)
	if err != nil {
		return err
	}

	nodes, err := s.nodeManager.ListNodes(cluster.Name)
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

		var nodeType node.Type
		switch {
		case strings.Contains(n.Name, string(node.Master)):
			nodeType = node.Master

		case strings.Contains(n.Name, string(node.Worker)):
			nodeType = node.Worker
		}

		if nodeType != "" {
			if err := s.join(n, nodeType, clusterDir); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *SkubaBootstrapper) DownloadKubeConfig(cluster *data.Cluster, destDir string) (string, error) {
	return downloadKubeConfig(s.nodeManager, cluster, "", destDir)
}

func (s *SkubaBootstrapper) Prepare(cluster *data.Cluster, force bool) error {
	return installSkubaExecutables(cluster.Spec.Version, force)
}

func (s *SkubaBootstrapper) Type() string {
	return constants.SKUBA
}

func (s *SkubaBootstrapper) init(cluster *data.Cluster, master *data.Node, extraOptions *SkubaExtraOptions) (string, error) {
	if err := s.register(cluster, extraOptions); err != nil {
		return "", err
	}

	cmds := []string{
		fmt.Sprintf("ssh-add %s", cluster.Spec.Prikey),
		fmt.Sprintf("skuba cluster init %s --control-plane %s -v 5", cluster.Name, master.Status.IPAddresses),
	}

	for _, c := range cmds {
		cmdArgs := strings.Split(c, " ")

		cmd := util.UpdateCommandDefaultLogWithInfo(
			exec.CommandContext(
				context.Background(),
				cmdArgs[0],
				cmdArgs[1:]...,
			),
		)
		cmd.Dir = cluster.Spec.LocalClusterDir()

		if err := cmd.Run(); err != nil {
			return "", errors.WithStack(err)
		}
	}

	return path.Join(cluster.Spec.LocalClusterDir(), cluster.Name), nil
}

func (s *SkubaBootstrapper) bootstrap(master *data.Node, clusterDir string, isSingleNode bool) error {
	cmds := []struct {
		cmdline string
		enabled bool
	}{
		{
			cmdline: fmt.Sprintf("skuba node bootstrap %s -t %s -v 5", master.Name, master.Status.IPAddresses),
			enabled: true,
		},
		{
			cmdline: fmt.Sprintf("KUBECONFIG=%s kubectl taint nodes --all node-role.kubernetes.io/master-", path.Join(clusterDir, "admin.conf")),
			enabled: isSingleNode,
		},
	}

	for _, c := range cmds {
		if !c.enabled {
			continue
		}

		cmdArgs := strings.Split(c.cmdline, " ")

		cmd := util.UpdateCommandDefaultLogWithInfo(
			exec.CommandContext(
				context.Background(),
				cmdArgs[0],
				cmdArgs[1:]...,
			),
		)
		cmd.Dir = clusterDir

		if err := cmd.Run(); err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

func (s *SkubaBootstrapper) join(node *data.Node, nodeType node.Type, clusterDir string) error {
	cmds := []string{
		fmt.Sprintf("skuba node join %s -t %s -r %s -v 5", node.Name, node.Status.IPAddresses, nodeType),
	}

	for _, c := range cmds {
		cmdArgs := strings.Split(c, " ")

		cmd := util.UpdateCommandDefaultLogWithInfo(
			exec.CommandContext(
				context.Background(),
				cmdArgs[0],
				cmdArgs[1:]...,
			),
		)
		cmd.Dir = clusterDir

		if err := cmd.Run(); err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

func (s *SkubaBootstrapper) register(cluster *data.Cluster, extraOptions *SkubaExtraOptions) error {
	nodes, err := s.nodeManager.ListNodes(cluster.Name)
	if err != nil {
		return err
	}

	wgDone := sync.WaitGroup{}
	wgDone.Add(len(nodes))

	chErr := make(chan error, len(nodes))
	defer close(chErr)

	for _, n := range nodes {
		go func(n *data.Node) {
			defer wgDone.Done()

			_ = retry.Do(func() error {
				sshClient, err := utilssh.NewClient(
					n.Name,
					cluster.Spec.Prikey,
					"root",
					n.Status.IPAddresses,
					nil,
				)
				if err != nil {
					chErr <- err
					return nil
				}

				cmds := []string{
					"swapoff -a",
					fmt.Sprintf("SUSEConnect -r %s", extraOptions.RegisterCode),
					"SUSEConnect -p sle-module-containers/15.1/x86_64",
					fmt.Sprintf("SUSEConnect -p caasp/4.0/x86_64 -r %s", extraOptions.RegisterCode),
				}

				for _, c := range cmds {
					err = sshClient.Run(nil, nil, c)
					if err != nil {
						chErr <- err
						return nil
					}
				}

				return nil
			},
				retry.Delay(10*time.Second),
				retry.MaxDelay(1*time.Minute),
			)
		}(n)
	}

	wgDone.Wait()
	err = nil

loop:
	for {
		select {
		case e := <-chErr:
			err = multierror.Append(err, e)

		default:
			break loop
		}
	}

	return err
}

func installSkubaExecutables(version string, force bool) error {
	scripts := []script.Type{
		script.InstallPrerequisitesSkuba,
	}

	for _, s := range scripts {
		if err := script.Download(s, config.TagVersion, force); err != nil {
			return err
		}

		if err := script.Run(s, config.TagVersion, func(cmd *exec.Cmd) error {
			cmd.Env = append(
				cmd.Env,
				config.SkubaVersionsEnvVars(version)...,
			)

			return nil
		}); err != nil {
			return err
		}
	}

	return nil
}
