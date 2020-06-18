package di

import (
	"github.com/innobead/kubefire/internal/config"
	"github.com/innobead/kubefire/pkg/bootstrap"
	"github.com/innobead/kubefire/pkg/cluster"
	pkgconfig "github.com/innobead/kubefire/pkg/config"
	"github.com/innobead/kubefire/pkg/node"
	"github.com/innobead/kubefire/pkg/output"
	"os"
	"sync"
)

var lock = &sync.Mutex{}

var (
	clusterManager cluster.Manager
	nodeManager    node.Manager
	configManager  pkgconfig.Manager
	bootstrapper   bootstrap.Bootstrapper
	outputer       output.Outputer
)

func Output() output.Outputer {
	lock.Lock()
	defer lock.Unlock()

	if outputer != nil {
		return outputer
	}

	outputType := output.DEFAULT

	switch config.Output {
	case "json":
		outputType = output.JSON

	case "yaml":
		outputType = output.YAML
	}

	if o, err := output.NewOutput(outputType, os.Stdout); err != nil {
		panic(err)
	} else {
		outputer = o
	}

	return outputer
}

func ClusterManager() cluster.Manager {
	nm := NodeManager()
	b := Bootstrapper()
	cm := ConfigManager()

	lock.Lock()
	defer lock.Unlock()

	if clusterManager != nil {
		return clusterManager
	}

	clusterManager = cluster.NewDefaultManager(
		nm,
		b,
		cm,
	)

	return clusterManager
}

func NodeManager() node.Manager {
	lock.Lock()
	defer lock.Unlock()

	if nodeManager != nil {
		return nodeManager
	}

	nodeManager = node.NewIgniteNodeManager()

	return nodeManager
}

func ConfigManager() pkgconfig.Manager {
	lock.Lock()
	defer lock.Unlock()

	if configManager != nil {
		return configManager
	}

	configManager = pkgconfig.NewLocalConfigManager()

	return configManager
}

func Bootstrapper() bootstrap.Bootstrapper {
	lock.Lock()
	defer lock.Unlock()

	if bootstrapper != nil {
		return bootstrapper
	}

	switch config.Bootstrap {
	case bootstrap.SKUBA:
		bootstrapper = bootstrap.NewSkubaBootstrapper()

	case bootstrap.KUBEADM:
		bootstrapper = bootstrap.NewKubeadmBootstrapper()
	}

	return bootstrapper
}
