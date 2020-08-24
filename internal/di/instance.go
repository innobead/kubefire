package di

import (
	"github.com/innobead/kubefire/internal/config"
	"github.com/innobead/kubefire/pkg/bootstrap"
	"github.com/innobead/kubefire/pkg/bootstrap/versionfinder"
	"github.com/innobead/kubefire/pkg/cluster"
	pkgconfig "github.com/innobead/kubefire/pkg/config"
	"github.com/innobead/kubefire/pkg/node"
	"github.com/innobead/kubefire/pkg/output"
	"os"
	"path"
	"reflect"
)

func Output() output.Outputer {
	objType := reflect.TypeOf(new(output.Outputer)).Elem()
	objKey := path.Join(objType.PkgPath(), objType.Name())

	obj := getObjFromContainer(objKey)
	if obj != nil {
		return obj.(output.Outputer)
	}

	outputType := output.DEFAULT
	switch config.Output {
	case "json":
		outputType = output.JSON
	case "yaml":
		outputType = output.YAML
	}

	outputer := output.NewOutput(outputType, os.Stdout)
	addObjToContainer(
		objKey,
		outputer,
	)

	return outputer
}

func ClusterManager() cluster.Manager {
	objType := reflect.TypeOf(new(cluster.Manager)).Elem()
	objKey := path.Join(objType.PkgPath(), objType.Name())

	obj := getObjFromContainer(objKey)
	if obj != nil {
		return obj.(cluster.Manager)
	}

	clusterManager := cluster.NewDefaultManager()
	addObjToContainer(
		objKey,
		clusterManager,
	)

	return clusterManager
}

func NodeManager() node.Manager {
	objType := reflect.TypeOf(new(node.Manager)).Elem()
	objKey := path.Join(objType.PkgPath(), objType.Name())

	obj := getObjFromContainer(objKey)
	if obj != nil {
		return obj.(node.Manager)
	}

	nodeManager := node.NewIgniteNodeManager()
	addObjToContainer(
		objKey,
		nodeManager,
	)

	return nodeManager
}

func ConfigManager() pkgconfig.Manager {
	objType := reflect.TypeOf(new(pkgconfig.Manager)).Elem()
	objKey := path.Join(objType.PkgPath(), objType.Name())

	obj := getObjFromContainer(objKey)
	if obj != nil {
		return obj.(pkgconfig.Manager)
	}

	configManager := pkgconfig.NewLocalConfigManager()
	addObjToContainer(
		objKey,
		configManager,
	)

	return configManager
}

func Bootstrapper() bootstrap.Bootstrapper {
	objType := reflect.TypeOf(new(bootstrap.Bootstrapper)).Elem()
	objKey := path.Join(objType.PkgPath(), objType.Name())

	obj := getObjFromContainer(objKey)
	if obj != nil {
		return obj.(bootstrap.Bootstrapper)
	}

	bootstrapper = bootstrap.New(config.Bootstrapper) // on purpose, because bootstrapper is required to initialize version finder at constructor
	addObjToContainer(
		objKey,
		bootstrapper,
	)

	return bootstrapper
}

func VersionFinder() versionfinder.Finder {
	objType := reflect.TypeOf(new(versionfinder.Finder)).Elem()
	objKey := path.Join(objType.PkgPath(), objType.Name())

	obj := getObjFromContainer(objKey)
	if obj != nil {
		return obj.(versionfinder.Finder)
	}

	versionFinder := versionfinder.New(bootstrapper.Type())
	addObjToContainer(
		objKey,
		versionFinder,
	)

	return versionFinder
}
