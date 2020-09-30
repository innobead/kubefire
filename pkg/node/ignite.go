package node

import (
	"bytes"
	"context"
	"fmt"
	"github.com/avast/retry-go"
	"github.com/innobead/kubefire/pkg/config"
	"github.com/innobead/kubefire/pkg/data"
	"github.com/innobead/kubefire/pkg/util"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"html/template"
	"os"
	"os/exec"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	RunCmd            = `ignite run {{.Image}} --name={{.Name}} --label=cluster={{.Cluster}} --ssh={{.Pubkey}} --kernel-image={{.KernelImage}} --kernel-image={{.KernelImage}} --cpus={{.Cpus}} --memory={{.Memory}} --size={{.DiskSize}}`
	CreateCmd         = `ignite create {{.Image}} --name={{.Name}} --label=cluster={{.Cluster}} --ssh={{.Pubkey}} --kernel-image={{.KernelImage}} --kernel-image={{.KernelImage}} --cpus={{.Cpus}} --memory={{.Memory}} --size={{.DiskSize}}`
	DeleteCmd         = "ignite rm {{.Name}} --force"
	StartCmd          = "ignite start {{.Name}}"
	StopCmd           = "ignite stop {{.Name}}"
	ListImageCmd      = "ignite {{.Image}} ls -q"
	InspectCmd        = "ignite inspect {{.Resource}} {{.ResourceName}} -t \"{{.ResourceFilter}}\""
	DeleteResourceCmd = "ignite {{.Resource}} rm {{.ResourceName}}"
)

type IgniteNodeManager struct {
}

type IgniteCache struct {
	Type        string
	Name        string
	Description string
}

func NewIgniteNodeManager() *IgniteNodeManager {
	return &IgniteNodeManager{}
}

func (i *IgniteNodeManager) CreateNodes(nodeType Type, node *config.Node, started bool) error {
	logrus.WithFields(logrus.Fields{
		"cluster": node.Cluster.Name,
		"started": started,
	}).Infof("creating %s nodes of cluster", nodeType)

	templateContent := CreateCmd
	if started {
		templateContent = RunCmd
	}

	tmp, err := template.New("create").Parse(templateContent)
	if err != nil {
		return errors.WithStack(err)
	}

	var wgCreateNode sync.WaitGroup

	for i := 1; i <= node.Count; i++ {
		tmpBuffer := &bytes.Buffer{}

		n := &struct {
			Name        string
			Cluster     string
			Image       string
			KernelImage string
			Pubkey      string
			Cpus        int
			Memory      string
			DiskSize    string
		}{
			Name:        Name(node.Cluster.Name, nodeType, i),
			Cluster:     node.Cluster.Name,
			Image:       node.Cluster.Image,
			KernelImage: node.Cluster.KernelImage,
			Pubkey:      node.Cluster.Pubkey,
			Cpus:        node.Cpus,
			Memory:      node.Memory,
			DiskSize:    node.DiskSize,
		}

		if err := tmp.Execute(tmpBuffer, n); err != nil {
			return errors.WithStack(err)
		}

		cmdArgs := strings.Split(tmpBuffer.String(), " ")
		cmdArgs = append(cmdArgs, fmt.Sprintf("--kernel-args=%s", node.Cluster.KernelArgs))

		cmd := util.UpdateCommandDefaultLogWithInfo(
			exec.CommandContext(
				context.Background(),
				"sudo",
				cmdArgs...,
			),
		)

		logrus.WithField("node", n.Name).Infoln("creating node")

		err := cmd.Start()
		if err != nil {
			return errors.WithStack(err)
		}

		wgCreateNode.Add(1)

		go func(name string) {
			defer wgCreateNode.Done()

			if err := cmd.Wait(); err != nil {
				logrus.WithField("node", name).WithError(err).Errorln("failed to create node")
			}
		}(n.Name)
	}

	wgCreateNode.Wait()

	return nil
}

func (i *IgniteNodeManager) DeleteNodes(nodeType Type, node *config.Node) error {
	logrus.WithField("cluster", node.Cluster.Name).Infof("deleting %s nodes", nodeType)

	for j := 1; j <= node.Count; j++ {
		name := Name(node.Cluster.Name, nodeType, j)
		if err := i.DeleteNode(name); err != nil {
			return err
		}
	}

	return nil
}

func (i *IgniteNodeManager) DeleteNode(name string) error {
	logrus.WithField("node", name).Infoln("deleting node")

	_, err := i.runCmd(
		"delete",
		DeleteCmd,
		struct {
			Name string
		}{
			Name: name,
		},
		true,
	)

	return err
}

func (i *IgniteNodeManager) GetNode(name string) (*data.Node, error) {
	logrus.WithField("node", name).Debugln("getting node")

	cmdArgs := strings.Split(fmt.Sprintf("ignite ps --all -f {{.ObjectMeta.Name}}=%s", name), " ")
	cmd := util.UpdateCommandDefaultLog(
		exec.CommandContext(
			context.Background(),
			"sudo",
			cmdArgs...,
		),
		logrus.DebugLevel,
	)

	if err := cmd.Run(); err != nil {
		return nil, err
	}

	node := &data.Node{
		Name:   name,
		Spec:   config.Node{Cluster: config.NewCluster()},
		Status: data.NodeStatus{},
	}

	nodeValueFilters := map[interface{}]map[string]string{
		node.Spec.Cluster: {
			"{{.ObjectMeta.Labels.cluster}}": "Name",
		},
		&node.Spec: {
			"{{.Spec.CPUs}}":     "Cpus",
			"{{.Spec.Memory}}":   "Memory",
			"{{.Spec.DiskSize}}": "DiskSize",
		},
		&node.Status: {
			"{{.Status.Running}}":     "Running",
			"{{.Status.IPAddresses}}": "IPAddresses",
			"{{.Status.Image.ID}}":    "Image",
			"{{.Status.Kernel.ID}}":   "Kernel",
		},
	}

	for v, filters := range nodeValueFilters {
		nodeValue := reflect.ValueOf(v).Elem()

		for filter, field := range filters {
			newCmdArgs := cmdArgs
			newCmdArgs = append(newCmdArgs, "-t "+filter)

			cmd := exec.CommandContext(context.Background(), "sudo", newCmdArgs...)

			logrus.Debugf("%+v", cmd.Args)

			output, err := cmd.Output()
			if err != nil {
				return nil, errors.WithStack(err)
			}

			if len(output) == 0 {
				return nil, errors.Errorf("%s node unavailable", name)
			}

			fieldValue := strings.TrimSuffix(strings.TrimSpace(string(output)), "\n")

			f := nodeValue.FieldByName(field)
			switch f.Kind() {
			case reflect.String:
				f.SetString(fieldValue)

			case reflect.Int:
				if i, err := strconv.ParseInt(fieldValue, 10, 64); err == nil {
					f.SetInt(i)
				}

			case reflect.Bool:
				if b, err := strconv.ParseBool(fieldValue); err == nil {
					f.SetBool(b)
				}
			}
		}
	}

	return node, nil
}

func (i *IgniteNodeManager) ListNodes(clusterName string) ([]*data.Node, error) {
	logrus.WithField("cluster", clusterName).Debugln("listing nodes of cluster")

	cmdArgs := strings.Split("ignite ps --all", " ")

	if clusterName != "" {
		cmdArgs = append(
			cmdArgs,
			"-f",
			fmt.Sprintf("{{.ObjectMeta.Name}}=~%s-", clusterName),
			"-t",
			"{{.ObjectMeta.Name}}",
		)
	}

	cmd := exec.CommandContext(context.Background(), "sudo", cmdArgs...)

	logrus.Debugf("%+v", cmd.Args)

	output, err := cmd.Output()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var nodes []*data.Node

	if len(output) > 0 {
		names := strings.Split(strings.TrimSpace(string(output)), "\n")

		for _, n := range names {
			if !IsValidNodeName(n, clusterName) {
				continue
			}

			node, err := i.GetNode(n)
			if err != nil {
				return nil, err
			}

			nodes = append(nodes, node)
		}
	}

	return nodes, nil
}

func (i *IgniteNodeManager) LoginBySSH(name string, configManager config.Manager) error {
	logrus.WithField("node", name).Infoln("ssh into node")

	node, err := i.GetNode(name)
	if err != nil {
		return err
	}

	cluster, err := configManager.GetCluster(node.Spec.Cluster.Name)
	if err != nil {
		return err
	}

	cmdArgs := strings.Split(fmt.Sprintf("ignite ssh -i %s %s", cluster.Prikey, name), " ")

	cmd := exec.CommandContext(
		context.Background(),
		"sudo",
		cmdArgs...,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (i *IgniteNodeManager) WaitNodesRunning(clusterName string, timeoutMin time.Duration) error {
	logrus.WithField("cluster", clusterName).Infoln("waiting nodes of cluster running")

	err := retry.Do(func() error {
		nodes, err := i.ListNodes(clusterName)
		if err != nil {
			return err
		}

		for _, n := range nodes {
			if !n.Status.Running {
				return errors.New(fmt.Sprintf("node (%s) is not running", n.Name))
			}
		}

		return nil
	}, retry.Delay(5*time.Second), retry.MaxDelay(timeoutMin*time.Minute))

	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (i *IgniteNodeManager) StartNodes(clusterName string) error {
	logrus.WithField("cluster", clusterName).Infoln("starting nodes of cluster running")

	nodes, err := i.ListNodes(clusterName)
	if err != nil {
		return err
	}

	for _, n := range nodes {
		if n.Status.Running {
			logrus.WithField("node", n.Name).Infoln("node is already running")
			continue
		}

		if err := i.StartNode(n.Name); err != nil {
			return err
		}
	}

	return nil
}

func (i *IgniteNodeManager) StartNode(name string) error {
	logrus.WithField("node", name).Infoln("starting node")

	_, err := i.runCmd(
		"start",
		StartCmd,
		struct {
			Name string
		}{
			Name: name,
		},
		true,
	)

	return err
}

func (i *IgniteNodeManager) StopNodes(clusterName string) error {
	logrus.WithField("cluster", clusterName).Infoln("stopping nodes of cluster running")

	nodes, err := i.ListNodes(clusterName)
	if err != nil {
		return err
	}

	for _, n := range nodes {
		if !n.Status.Running {
			logrus.WithField("node", n.Name).Infoln("node is already stopped")
			continue
		}

		if err := i.StopNode(n.Name); err != nil {
			return err
		}
	}

	return nil
}

func (i *IgniteNodeManager) StopNode(name string) error {
	logrus.WithField("node", name).Infoln("stopping node")

	_, err := i.runCmd(
		"stop",
		StopCmd,
		struct {
			Name string
		}{
			Name: name,
		},
		true,
	)

	return err
}

func (i *IgniteNodeManager) GetCaches() ([]interface{}, error) {
	var caches []interface{}

	resources := []struct {
		Image          string
		Filters        []string
		Resource       string
		ResourceName   string
		ResourceFilter string
	}{
		{
			Image:          "image",
			Filters:        []string{"{{.Spec.OCI}}", "{{.Status.OCISource.ID}}"},
			Resource:       "image",
			ResourceFilter: "{{.Name}}",
		},
		{
			Image:          "kernel",
			Filters:        []string{"{{.Status.OCISource.ID}}"},
			Resource:       "kernel",
			ResourceFilter: "{{.Name}}",
		},
	}

	for _, resource := range resources {
		var cache IgniteCache

		output, err := i.runCmd("", ListImageCmd, resource, false)
		if err != nil {
			return nil, err
		}
		if output == "" {
			continue
		}

		imgIds := strings.Split(output, "\n")

		for _, imgId := range imgIds {
			imgId = strings.Trim(imgId, "\t\n")

			c := resource
			c.ResourceName = imgId

			imgName, err := i.runCmd("", InspectCmd, c, false)
			if err != nil {
				return nil, err
			}

			var imgDescription []string
			for _, f := range c.Filters {
				c.ResourceFilter = f
				description, err := i.runCmd("", InspectCmd, c, false)
				if err != nil {
					return nil, err
				}

				imgDescription = append(imgDescription, strings.Trim(description, `"`))
			}

			cache.Type = c.Image
			cache.Name = strings.Trim(imgName, `"`)
			cache.Description = strings.Join(imgDescription, ",")
		}

		caches = append(caches, &cache)
	}

	return caches, nil
}

func (i *IgniteNodeManager) DeleteCaches() error {
	logrus.Infof("Deleting ignite image caches")

	caches, err := i.GetCaches()
	if err != nil {
		return err
	}

	for _, c := range caches {
		c := c.(*IgniteCache)

		_, err := i.runCmd(
			"",
			DeleteResourceCmd,
			struct {
				Resource     string
				ResourceName string
			}{
				Resource:     c.Type,
				ResourceName: c.Name,
			},
			true,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (i *IgniteNodeManager) runCmd(templateName string, templateContent string, templateVars interface{}, logOutput bool) (string, error) {
	tmp, err := template.New(templateName).Parse(templateContent)
	if err != nil {
		return "", errors.WithStack(err)
	}

	tmpBuffer := &bytes.Buffer{}
	if err := tmp.Execute(tmpBuffer, templateVars); err != nil {
		return "", errors.WithStack(err)
	}

	cmd := exec.CommandContext(
		context.Background(),
		"sudo",
		strings.Split(tmpBuffer.String(), " ")...,
	)

	if logOutput {
		cmd = util.UpdateCommandDefaultLogWithInfo(cmd)

		if err := cmd.Run(); err != nil {
			return "", errors.WithStack(err)
		}

		return "", nil
	}

	outputBuffer := &bytes.Buffer{}
	cmd.Stdout = outputBuffer
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return "", errors.WithStack(err)
	}

	return strings.Trim(outputBuffer.String(), "\t\n"), nil
}
