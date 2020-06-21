package node

import (
	"fmt"
	"github.com/innobead/kubefire/pkg/config"
	"github.com/innobead/kubefire/pkg/data"
	"strconv"
	"time"
)

type Type string

const (
	Admin  Type = "admin"
	Master Type = "master"
	Worker Type = "worker"
)

type Manager interface {
	CreateNodes(nodeType Type, node *config.Node) error
	DeleteNodes(nodeType Type, node *config.Node) error
	DeleteNode(name string) error
	GetNode(name string) (*data.Node, error)
	ListNodes(clusterName string) ([]*data.Node, error)
	LoginBySSH(name string, configManager config.Manager) error
	WaitNodesRunning(clusterName string, timeoutMin time.Duration) error
}

func NodeName(clusterName string, nodeType Type, index int) string {
	return fmt.Sprintf("%s-%s-%s", clusterName, nodeType, strconv.Itoa(index))
}
