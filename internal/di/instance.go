package di

import (
	"github.com/innobead/kubefire/internal/config"
	"github.com/innobead/kubefire/pkg/bootstrap"
	"github.com/innobead/kubefire/pkg/bootstrap/versionfinder"
	"github.com/innobead/kubefire/pkg/cache"
	"github.com/innobead/kubefire/pkg/cluster"
	pkgconfig "github.com/innobead/kubefire/pkg/config"
	"github.com/innobead/kubefire/pkg/node"
	"github.com/innobead/kubefire/pkg/output"
	"os"
)

func Output() output.Outputer {
	return addObjToContainer(
		new(output.Outputer),
		func() interface{} {
			outputType := output.DEFAULT
			switch config.Output {
			case "json":
				outputType = output.JSON
			case "yaml":
				outputType = output.YAML
			}

			return output.NewOutput(outputType, os.Stdout)
		},
	).(output.Outputer)
}

func ClusterManager() cluster.Manager {
	return addObjToContainer(
		new(cluster.Manager),
		func() interface{} {
			return cluster.NewDefaultManager()
		},
	).(cluster.Manager)
}

func NodeManager() node.Manager {
	return addObjToContainer(
		new(node.Manager),
		func() interface{} {
			return node.NewIgniteNodeManager()
		},
	).(node.Manager)
}

func ConfigManager() pkgconfig.Manager {
	return addObjToContainer(
		new(pkgconfig.Manager),
		func() interface{} {
			return pkgconfig.NewLocalConfigManager()
		},
	).(pkgconfig.Manager)
}

func Bootstrapper() bootstrap.Bootstrapper {
	addObjToContainer(
		new(bootstrap.Bootstrapper),
		func() interface{} {
			bootstrapper = bootstrap.New(config.Bootstrapper) // on purpose, because bootstrapper is required to initialize version finder at constructor
			return bootstrapper
		},
	)

	return bootstrapper
}

func VersionFinder() versionfinder.Finder {
	return addObjToContainer(
		new(versionfinder.Finder),
		func() interface{} {
			return versionfinder.New(bootstrapper.Type())
		},
	).(versionfinder.Finder)
}

func CacheManager() cache.Manager {
	return addObjToContainer(
		new(cache.Manager),
		func() interface{} {
			return cache.NewLocalManager(pkgconfig.RootDir)
		},
	).(cache.Manager)
}
