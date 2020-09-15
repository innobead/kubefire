package config

import (
	"encoding/json"
	"github.com/innobead/kubefire/pkg/constants"
	"github.com/pkg/errors"
	"path"
	"regexp"
	"strings"
)

type Cluster struct {
	Name         string `json:"name"`
	Bootstrapper string `json:"bootstrapper"`
	Pubkey       string `json:"pubkey"`
	Prikey       string `json:"prikey"`
	Version      string `json:"version"`

	Image       string `json:"image"`
	KernelImage string `json:"kernel_image,omitempty"`
	KernelArgs  string `json:"kernel_args,omitempty"`

	ExtraOptions map[string]interface{} `json:"extra_options"`
	Deployed     bool                   `json:"deployed"` // the only status property

	Master Node `json:"master"`
	Worker Node `json:"worker"`
}

func NewCluster() *Cluster {
	c := Cluster{
		Master:       Node{},
		Worker:       Node{},
		ExtraOptions: map[string]interface{}{},
	}

	c.Master.Cluster = &c
	c.Worker.Cluster = &c

	return &c
}

func NewDefaultCluster() *Cluster {
	cluster := NewCluster()

	cluster.Bootstrapper = constants.KUBEADM
	cluster.Image = "ghcr.io/innobead/kubefire-opensuse-leap:15.2"
	cluster.KernelImage = "ghcr.io/innobead/kubefire-ignite-kernel:4.19.125-amd64"
	cluster.KernelArgs = "console=ttyS0 reboot=k panic=1 pci=off ip=dhcp security=apparmor apparmor=1"
	cluster.Master.Count = 1
	cluster.Master.Cpus = 2
	cluster.Master.Memory = "2GB"
	cluster.Master.DiskSize = "10GB"
	cluster.Worker.Count = 0
	cluster.Worker.Cpus = 2
	cluster.Worker.Memory = "2GB"
	cluster.Worker.DiskSize = "10GB"

	return cluster
}

type Node struct {
	Count    int      `json:"count"`
	Memory   string   `json:"memory,omitempty"`
	Cpus     int      `json:"cpus,omitempty"`
	DiskSize string   `json:"disk_size,omitempty"`
	Cluster  *Cluster `json:"-"`
}

func (c *Cluster) LocalClusterDir() string {
	return path.Join(ClusterRootDir, c.Name)
}

func (c *Cluster) LocalKubeConfig() string {
	return path.Join(c.LocalClusterDir(), "admin.conf")
}

func (c *Cluster) LocalClusterConfigFile() string {
	return path.Join(c.LocalClusterDir(), "cluster.yaml")
}

func (c *Cluster) LocalClusterKeyFiles() (string, string) {
	return path.Join(c.LocalClusterDir(), "key"), path.Join(c.LocalClusterDir(), "key.pub")
}

func (c *Cluster) UpdateExtraOptions(options string) {
	if options == "" {
		return
	}

	optionList := strings.Split(options, " ")

	for _, option := range optionList {
		values := strings.SplitN(option, "=", 2)

		if len(values) != 2 {
			continue
		}

		if strings.Contains(values[1], "=") {
			pattern := regexp.MustCompile(`^['"]?(\S+)['"]?$`)
			matches := pattern.FindStringSubmatch(values[1])

			if len(matches) == 2 {
				if _, ok := c.ExtraOptions[values[0]]; !ok {
					c.ExtraOptions[values[0]] = []string{}
				}

				c.ExtraOptions[values[0]] = append(c.ExtraOptions[values[0]].([]string), strings.Split(matches[1], ",")...)
			}

			continue
		}

		c.ExtraOptions[values[0]] = values[1]
	}
}

func (c *Cluster) ParseExtraOptions(options interface{}) error {
	bytes, err := json.Marshal(c.ExtraOptions)
	if err != nil {
		return errors.WithStack(err)
	}

	err = json.Unmarshal(bytes, options)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
