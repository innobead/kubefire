package di

import (
	"github.com/innobead/kubefire/pkg/bootstrap"
	"github.com/innobead/kubefire/pkg/bootstrap/versionfinder"
	"github.com/innobead/kubefire/pkg/cache"
	"github.com/innobead/kubefire/pkg/cluster"
	pkgconfig "github.com/innobead/kubefire/pkg/config"
	"github.com/innobead/kubefire/pkg/node"
	"github.com/innobead/kubefire/pkg/output"
)

type BootstrapperAware interface {
	SetBootstrapper(bootstrapper bootstrap.Bootstrapper)
}

type VersionFinderAware interface {
	SetVersionFinder(versionFinder versionfinder.Finder)
}

type ConfigManagerAware interface {
	SetConfigManager(configManager pkgconfig.Manager)
}

type ClusterManagerAware interface {
	SetClusterManager(clusterManager cluster.Manager)
}

type NodeManagerAware interface {
	SetNodeManager(nodeManager node.Manager)
}

type OutputAware interface {
	SetOutputer(outputer output.Outputer)
}

type CacheManagerAware interface {
	SetCacheManager(cacheManager cache.Manager)
}
