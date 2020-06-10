package node

import (
	"bytes"
	"fmt"
	"github.com/innobead/kubefire/pkg/config"
	"github.com/innobead/kubefire/pkg/data"
	"github.com/sirupsen/logrus"
	"html/template"
	"os"
	"os/exec"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

const (
	RunCmd    = "run {{.Image}} --name={{.Name}} --ssh --kernel-image={{.KernelImage}} --cpus={{.Cpus}} --memory={{.Memory}} --size={{.DiskSize}}"
	DeleteCmd = "rm {{.Name}} --force"
)

type IgniteNodeManager struct {
}

func NewIgniteNodeManager() *IgniteNodeManager {
	return &IgniteNodeManager{}
}

func (i *IgniteNodeManager) CreateNodes(nodeType Type, node *config.Node) Error {
	t, err := template.New("create").Parse(RunCmd)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup

	for i := 1; i <= node.Count; i++ {
		buf := &bytes.Buffer{}
		c := &struct {
			Name        string
			Image       string
			KernelImage string
			KernelArgs  string
			Cpus        int
			Memory      string
			DiskSize    string
		}{
			Name:        fmt.Sprintf("%s-%s-%s", node.Cluster.Name, nodeType, strconv.Itoa(i)),
			Image:       node.Cluster.Image,
			KernelImage: node.Cluster.KernelImage,
			KernelArgs:  node.Cluster.KernelArgs,
			Cpus:        node.Cpus,
			Memory:      node.Memory,
			DiskSize:    node.DiskSize,
		}

		if err := t.Execute(buf, c); err != nil {
			return err
		}

		cmd := exec.Command("ignite", strings.Split(buf.String(), " ")...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err := cmd.Start()
		if err != nil {
			return err
		}

		wg.Add(1)

		go func() {
			defer wg.Done()

			if err := cmd.Wait(); err != nil {
				logrus.WithError(err).Errorln("")
			}
		}()
	}

	wg.Wait()

	return nil
}

func (i *IgniteNodeManager) DeleteNodes(nodeType Type, node *config.Node) Error {
	t, err := template.New("create").Parse(DeleteCmd)
	if err != nil {
		return err
	}

	for i := 1; i <= node.Count; i++ {
		buf := &bytes.Buffer{}
		c := &struct {
			Name string
		}{
			Name: fmt.Sprintf("%s-%s-%s", node.Cluster.Name, nodeType, strconv.Itoa(i)),
		}

		if err := t.Execute(buf, c); err != nil {
			return err
		}

		cmd := exec.Command("ignite", strings.Split(buf.String(), " ")...)
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	return nil
}

func (i *IgniteNodeManager) Delete(name string) Error {
	t, err := template.New("delete").Parse(DeleteCmd)
	if err != nil {
		return err
	}

	buf := &bytes.Buffer{}
	c := &struct {
		Name string
	}{
		Name: name,
	}
	if err := t.Execute(buf, c); err != nil {
		return err
	}

	cmd := exec.Command("ignite", strings.Split(buf.String(), " ")...)
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func (i *IgniteNodeManager) Get(name string) (*data.Node, Error) {
	args := strings.Split(fmt.Sprintf("ps --all -f {{.ObjectMeta.Name}}=%s", name), " ")
	cmd := exec.Command("ignite", args...)

	if err := cmd.Run(); err != nil {
		return nil, err
	}

	node := &data.Node{
		Name:   name,
		Spec:   config.Node{},
		Status: data.NodeStatus{},
	}

	nodeValueFilters := map[interface{}]map[string]string{
		&node.Spec: {
			"{{.Spec.CPUs}}":     "Cpus",
			"{{.Spec.Memory}}":   "Memory",
			"{{.Spec.DiskSize}}": "DiskSize",
		},
		&node.Status: {
			"{{.Status.Running}}": "Running",
		},
	}

	for v, filters := range nodeValueFilters {
		nodeValue := reflect.ValueOf(v).Elem()

		for filter, field := range filters {
			cmdArgs := args
			cmdArgs = append(cmdArgs, "-t "+filter)

			cmd := exec.Command("ignite", cmdArgs...)
			output, err := cmd.Output()
			if err != nil {
				return nil, err
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

func (i *IgniteNodeManager) List(clusterName string) ([]*data.Node, Error) {
	args := strings.Split("ps --all", " ")

	var cmd *exec.Cmd
	if clusterName != "" {
		args = append(
			args,
			"-f",
			fmt.Sprintf("{{.ObjectMeta.Name}}=~%s", clusterName),
			"-t",
			"{{.ObjectMeta.Name}}",
		)
	}

	cmd = exec.Command("ignite", args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var nodes []*data.Node

	if len(output) > 0 {
		names := strings.Split(strings.TrimSpace(string(output)), "\n")
		for _, n := range names {
			node, err := i.Get(n)
			if err != nil {
				return nil, err
			}

			nodes = append(nodes, node)
		}
	}

	return nodes, nil
}
