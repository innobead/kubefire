package config

import (
	"reflect"
	"strconv"
	"strings"
)

type Cluster struct {
	Name         string `json:"name"`
	Bootstrapper string `json:"bootstrapper"`
	Pubkey       string `json:"pubkey"`
	Prikey       string `json:"prikey"`

	Image       string `json:"image"`
	KernelImage string `json:"kernel_image,omitempty"`
	KernelArgs  string `json:"kernel_args,omitempty"`

	Admin  Node `json:"admin"`
	Master Node `json:"master"`
	Worker Node `json:"worker"`

	ExtraOptions string `json:"extra_options"`
}

type Node struct {
	Count    int      `json:"count"`
	Memory   string   `json:"memory,omitempty"`
	Cpus     int      `json:"cpus,omitempty"`
	DiskSize string   `json:"disk_size,omitempty"`
	Cluster  *Cluster `json:"-"`
}

func (c *Cluster) ParseExtraOptions(obj interface{}) interface{} {
	value := reflect.ValueOf(obj).Elem()

	optionList := strings.Split(c.ExtraOptions, ",")

	for _, option := range optionList {
		values := strings.Split(option, "=")
		if len(values) == 2 {
			field := value.FieldByName(values[0])

			switch field.Kind() {
			case reflect.String:
				field.SetString(values[1])

			case reflect.Int:
				v, _ := strconv.Atoi(values[1])
				field.SetInt(int64(v))

			case reflect.Bool:
				b, _ := strconv.ParseBool(values[1])
				field.SetBool(b)
			}
		}
	}

	return value.Interface()
}
