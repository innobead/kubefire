package bootstrap

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-multierror"
	"github.com/innobead/kubefire/pkg/config"
	"github.com/innobead/kubefire/pkg/data"
	"github.com/innobead/kubefire/pkg/node"
	"github.com/innobead/kubefire/pkg/util/ssh"
	"github.com/pkg/errors"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"
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
	extraOptions := cluster.Spec.ParseExtraOptions(SkubaExtraOptions{}).(SkubaExtraOptions)

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
		cmd := exec.CommandContext(context.Background(), c)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		cmd.Dir = config.LocalClusterDir(cluster.Name)

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

		cmd := exec.CommandContext(context.Background(), c.cmdline)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
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
		cmd := exec.CommandContext(context.Background(), c)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
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

			client, err := ssh.NewClient(n.Name, cluster.Spec.Prikey, "root", n.Status.IPAddresses, nil)
			if err != nil {
				chErr <- err
				return
			}

			err = client.Run(nil, nil, fmt.Sprintf("SUSEConnect -r %s", extraOptions.RegisterCode))
			if err != nil {
				chErr <- err
				return
			}
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
