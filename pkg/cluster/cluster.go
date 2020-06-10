package cluster

import (
	"github.com/innobead/kubefire/pkg/bootstrap"
	pkgconfig "github.com/innobead/kubefire/pkg/config"
	"github.com/innobead/kubefire/pkg/data"
	"github.com/innobead/kubefire/pkg/node"
	"github.com/sirupsen/logrus"
)

type Error error

type Manager interface {
	Init(cluster *pkgconfig.Cluster) Error
	Create(name string) Error
	Delete(name string, force bool) Error
	Get(name string) (*data.Cluster, Error)
	List() ([]*data.Cluster, Error)
}

type DefaultManager struct {
	NodeManager   node.Manager
	Bootstrapper  bootstrap.Bootstrapper
	ConfigManager pkgconfig.Manager
}

func NewDefaultManager(nodeManager node.Manager, bootstrapper bootstrap.Bootstrapper, configManager pkgconfig.Manager) (Manager, error) {
	if nodeManager == nil {
		nodeManager = node.NewIgniteNodeManager()
	}

	if bootstrapper == nil {
		bootstrapper = bootstrap.NewKubeadmBootstrapper()
	}

	if configManager == nil {
		configManager = pkgconfig.NewLocalConfigManager()
	}

	return &DefaultManager{
		NodeManager:   nodeManager,
		Bootstrapper:  bootstrapper,
		ConfigManager: configManager,
	}, nil
}

func (d *DefaultManager) Init(cluster *pkgconfig.Cluster) Error {
	if _, err := d.ConfigManager.Get(cluster.Name); err == nil {
		logrus.Warnf("cluster (%s) config (%s) already exists", cluster.Name, pkgconfig.ClusterConfigFile(cluster.Name))
	}

	return d.ConfigManager.Save(cluster.Name, cluster)
}

func (d *DefaultManager) Create(name string) Error {
	cluster, err := d.ConfigManager.Get(name)
	if err != nil {
		return err
	}

	nodeTypeConfigs := map[node.Type]*pkgconfig.Node{
		node.Admin:  &cluster.Admin,
		node.Master: &cluster.Master,
		node.Worker: &cluster.Worker,
	}
	for t, c := range nodeTypeConfigs {
		if err := d.NodeManager.CreateNodes(t, c); err != nil {
			return err
		}
	}

	return nil
}

func (d *DefaultManager) Delete(name string, force bool) Error {
	cluster, err := d.ConfigManager.Get(name)
	if err != nil {
		return err
	}

	nodeTypeConfigs := map[node.Type]*pkgconfig.Node{
		node.Admin:  &cluster.Admin,
		node.Master: &cluster.Master,
		node.Worker: &cluster.Worker,
	}
	for t, c := range nodeTypeConfigs {
		if err := d.NodeManager.DeleteNodes(t, c); err != nil {
			if !force {
				return err
			} else {
				logrus.WithError(err).Warnln("failed to delete nodes")
			}
		}
	}

	if err := d.ConfigManager.Delete(name); err != nil {
		return err
	}

	return nil
}

func (d *DefaultManager) Get(name string) (*data.Cluster, Error) {
	c, err := d.ConfigManager.Get(name)
	if err != nil {
		return nil, err
	}

	ns, err := d.NodeManager.List(name)
	if err != nil {
		return nil, err
	}

	return &data.Cluster{
		Name:  c.Name,
		Spec:  *c,
		Nodes: ns,
	}, nil
}

func (d *DefaultManager) List() ([]*data.Cluster, Error) {
	cs, err := d.ConfigManager.List()
	if err != nil {
		return nil, err
	}

	var clusters []*data.Cluster

	for _, c := range cs {
		clusters = append(clusters, &data.Cluster{
			Name: c.Name,
			Spec: *c,
		})
	}

	return clusters, nil
}
