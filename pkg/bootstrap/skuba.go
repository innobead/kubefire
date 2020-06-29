package bootstrap

import (
	"context"
	"fmt"
	"github.com/avast/retry-go"
	"github.com/hashicorp/go-multierror"
	"github.com/innobead/kubefire/pkg/config"
	"github.com/innobead/kubefire/pkg/data"
	"github.com/innobead/kubefire/pkg/node"
	"github.com/innobead/kubefire/pkg/util"
	"github.com/innobead/kubefire/pkg/util/ssh"
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

func (s *SkubaBootstrapper) Deploy(cluster *data.Cluster) error {
	extraOptions := cluster.Spec.ParseExtraOptions(&SkubaExtraOptions{}).(SkubaExtraOptions)

	if err := s.nodeManager.WaitNodesRunning(cluster.Name, 5); err != nil {
		return errors.WithMessage(err, "some nodes are not running")
	}

	firstMaster, err := s.nodeManager.GetNode(node.NodeName(cluster.Name, node.Master, 1))
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

		cmd := exec.CommandContext(
			context.Background(),
			cmdArgs[0],
			cmdArgs[1:]...,
		)
		cmd.Dir = config.LocalClusterDir(cluster.Name)
		util.UpdateDefaultCmdPipes(cmd)

		if err := cmd.Run(); err != nil {
			return "", errors.WithStack(err)
		}
	}

	return path.Join(config.LocalClusterDir(cluster.Name), cluster.Name), nil
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

		cmd := exec.CommandContext(context.Background(), cmdArgs[0], cmdArgs[1:]...)
		cmd.Dir = clusterDir
		util.UpdateDefaultCmdPipes(cmd)

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

		cmd := exec.CommandContext(context.Background(), cmdArgs[0], cmdArgs[1:]...)
		cmd.Dir = clusterDir
		util.UpdateDefaultCmdPipes(cmd)

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
				sshClient, err := ssh.NewClient(
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

				client, err := ssh.NewClient(n.Name, cluster.Spec.Prikey, "root", n.Status.IPAddresses, nil)
				if err != nil {
					chErr <- err
					return nil
				}

				cmds := []string{
					fmt.Sprintf("SUSEConnect -r %s", extraOptions.RegisterCode),
					"SUSEConnect -p sle-module-containers/15.1/x86_64",
					fmt.Sprintf("SUSEConnect -p caasp/4.0/x86_64 -r %s", extraOptions.RegisterCode),
				}

				for _, c := range cmds {
					err = client.Run(nil, nil, c)
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
