package config

import "github.com/innobead/kubefire/pkg/bootstrap"

type Cluster struct {
	Nodes        Nodes          `json:"nodes"`
	Name         string         `json:"name"`
	Pubkey       string         `json:"pubkey"`
	Bootstrapper bootstrap.Type `json:"bootstrapper"`
}

type Nodes struct {
	Count       int      `json:"count"`
	Image       string   `json:"image"`
	Pubkey      string   `json:"pubkey,omitempty"`
	CopyFiles   []string `json:"copy_files,omitempty"`
	KernelImage string   `json:"kernel_image,omitempty"`
	KernelArgs  string   `json:"kernel_args,omitempty"`
	Memory      string   `json:"memory,omitempty"`
	Cpus        int      `json:"cpus,omitempty"`
	Size        string   `json:"size,omitempty"`
}
