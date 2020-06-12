package util

import (
	"github.com/innobead/kubefire/internal/config"
	"github.com/innobead/kubefire/pkg/cluster"
	"github.com/innobead/kubefire/pkg/output"
	"os"
	"sync"
)

var lock = &sync.Mutex{}

var (
	clusterManager cluster.Manager
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
	lock.Lock()
	defer lock.Unlock()

	if clusterManager != nil {
		return clusterManager
	}

	if c, err := cluster.NewDefaultManager(nil, nil, nil); err != nil {
		panic(err)
	} else {
		clusterManager = c
	}

	return clusterManager
}
