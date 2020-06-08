package node

import (
	"bytes"
	"fmt"
	"github.com/innobead/kubefire/pkg/config"
	"github.com/innobead/kubefire/pkg/data"
	"github.com/sirupsen/logrus"
	"html/template"
	"os/exec"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

const (
	CreateCmd   = "create {{.Image}} --name={{.Name}} --ssh --kernel-args=\"{{.KernelArgs}}\" --kernel-image=\"{{.KernelImage}}\" --cpus={{.Cpus}} --memory={{.Memory}} --size={{.DiskSize}}"
	DeleteCmd   = "rm {{.Name}} --force"
	PsByNameCmd = "ps -f \"`{{.ObjectMeta.Name}}`={{.Name}}\""
)

type IgniteNodeManager struct {
}

func NewIgniteNodeManager() *IgniteNodeManager {
	return &IgniteNodeManager{}
}

func (i *IgniteNodeManager) CreateNodes(clusterName string, node *config.Node) Error {
	t, err := template.New("create").Parse(CreateCmd)
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
			Size        string
		}{
			Name:        fmt.Sprintf("%s-%s", clusterName, string(i)),
			Image:       node.Cluster.Image,
			KernelImage: node.Cluster.KernelImage,
			KernelArgs:  node.Cluster.KernelArgs,
			Cpus:        node.Cpus,
			Memory:      node.Memory,
			Size:        node.DiskSize,
		}

		if err := t.Execute(buf, c); err != nil {
			return err
		}

		cmd := exec.Command("ignite", strings.Split(buf.String(), " ")...)
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

func (i *IgniteNodeManager) DeleteNodes(clusterName string, node *config.Node) Error {
	t, err := template.New("create").Parse(DeleteCmd)
	if err != nil {
		return err
	}

	for i := 1; i <= node.Count; i++ {
		buf := &bytes.Buffer{}
		c := &struct {
			Name string
		}{
			Name: fmt.Sprintf("%s-%s", clusterName, string(i)),
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
	t, err := template.New("get").Parse(PsByNameCmd)
	if err != nil {
		return nil, err
	}

	buf := &bytes.Buffer{}
	c := &struct {
		Name string
	}{
		Name: name,
	}
	if err := t.Execute(buf, c); err != nil {
		return nil, err
	}

	args := strings.Split(buf.String(), " ")
	cmd := exec.Command("ignite", args...)
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	node := &data.Node{
		Name:   name,
		Config: &config.Node{},
	}
	nodeValue := reflect.ValueOf(&node.Config).Elem()

	filters := map[string]string{
		"{{.Spec.Image.OCI}}":      "Image",
		"{{.Spec.Kernel.OCI}}":     "KernelImage",
		"{{.Spec.Kernel.CmdLine}}": "KernelArgs",
		"{{.Spec.CPUs}}":           "Cpus",
		"{{.Spec.Memory}}":         "Memory",
		"{{.Spec.DiskSize}}":       "DiskSize",
	}
	for filter, field := range filters {
		cmdArgs := args
		cmdArgs = append(cmdArgs, filter)

		cmd := exec.Command("ignite", cmdArgs...)
		if err := cmd.Run(); err != nil {
			return nil, err
		}

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
			if i, err := strconv.ParseInt(fieldValue, 10, 64); err != nil {
				f.SetInt(i)
			}
		}
	}

	return node, nil
}

func (i *IgniteNodeManager) List() ([]*data.Node, Error) {
	cmd := exec.Command("ignite", "ps", "--all", "-t", "{{.ObjectMeta.Name}}")
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var nodes []*data.Node

	names := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, n := range names {
		nodes = append(nodes, &data.Node{Name: n})
	}

	return nodes, nil
}
