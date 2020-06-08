package config

type Cluster struct {
	Name         string `json:"name"`
	Bootstrapper string `json:"bootstrapper"`
	Pubkey       string `json:"pubkey"`

	Image       string `json:"image"`
	KernelImage string `json:"kernel_image,omitempty"`
	KernelArgs  string `json:"kernel_args,omitempty"`

	Admin  Node `json:"admin"`
	Master Node `json:"master"`
	Worker Node `json:"worker"`
}

type Node struct {
	Count    int      `json:"count"`
	Memory   string   `json:"memory,omitempty"`
	Cpus     int      `json:"cpus,omitempty"`
	DiskSize string   `json:"disk_size,omitempty"`
	Cluster  *Cluster `json:"-"`
}
